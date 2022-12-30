package listener

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"listener_cache_writethrough/src/config"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
)

const ttl = time.Duration(600 * time.Second) // 10 minutes

type Listener interface {
	InitializeCache(config.Config)
	InitializeListener(cfg config.Config)
	Start()
	GetCache() *redis.Client
	CloseDB()

	waitForNotification(*pq.Listener)
}

type ListenerImpl struct {
	ctx          context.Context
	Cache        *redis.Client
	ListenerConn *pq.Listener
}

type Record struct {
	Id          string `json:"id"`
	CreatedDate string `json:"created_date"`
}

func New(cfg config.Config, ctx context.Context) Listener {

	lstnr := &ListenerImpl{ctx: ctx}
	lstnr.InitializeCache(cfg)
	lstnr.InitializeListener(cfg)

	return lstnr
}

func (listener *ListenerImpl) waitForNotification(l *pq.Listener) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	doneCh := make(chan struct{})

	go func() {
		for {
			select {
			case n := <-l.Notify:

				// for shutdown
				if n == nil {
					return
				}

				var prettyJSON bytes.Buffer
				err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")
				if err != nil {
					fmt.Println("Error processing JSON: ", err)
					return
				}

				var record Record
				json.Unmarshal([]byte(string(n.Extra)), &record)

				listener.Cache.Set(listener.ctx, record.Id, record.CreatedDate, ttl).Result()

				fmt.Printf("\nLoading into cache\n\n%s", prettyJSON.String())

				return
			case <-time.After(90 * time.Second):
				fmt.Println("Received no events for 90 seconds, checking connection")
				go func() {
					l.Ping()
				}()
				return

			case <-sigchan:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	listener.CloseDB()
}

func (l *ListenerImpl) InitializeCache(cfg config.Config) {

	connstr := cfg.GetCacheConnDetails()

	rdb := redis.NewClient(&redis.Options{
		Addr:     connstr.Addr,
		Password: connstr.Password,
		DB:       0,
	})

	ping, _ := rdb.Ping(l.ctx).Result()

	if ping != "PONG" {
		panic(ping)
	}

	fmt.Println("connected to cache")

	l.Cache = rdb
}

func (l *ListenerImpl) InitializeListener(cfg config.Config) {
	dns := cfg.Dns()

	_, err := sql.Open("postgres", dns)
	if err != nil {
		log.Fatalf("failed to start db instance: %s", err.Error())
	}

	fmt.Println("connected to db")

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Fatalf("failed to create listener: %s", err.Error())
		}
	}

	l.ListenerConn = pq.NewListener(dns, 10*time.Second, time.Minute, reportProblem)
}

func (l *ListenerImpl) Start() {
	err := l.ListenerConn.Listen("row_inserted")
	if err != nil {
		panic(err)
	}

	fmt.Println("Start monitoring PostgreSQL...")

	for {
		l.waitForNotification(l.ListenerConn)
	}
}

func (l *ListenerImpl) GetCache() *redis.Client {
	return l.Cache
}

func (l *ListenerImpl) CloseDB() {
	if err := l.ListenerConn.Close(); err != nil {
		log.Fatalf("failed to close DB: %s", err.Error())
	} else {
		l.ListenerConn.Unlisten("row_inserted")
		log.Println("db connection closed")
		os.Exit(0)
	}
}
