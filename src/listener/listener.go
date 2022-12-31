package listener

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"listener_cache_writethrough/src/config"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
)

const ttl = time.Duration(600 * time.Second) // 10 minutes

type Listener interface {
	InitializeCache(config.Config) error
	InitializeListener(config.Config)
	Start(time.Duration) error
	GetCache() IRedis
	CloseDB() error
	CloseChan()
	sendErr(err error)

	waitForNotification(IListener, time.Duration)
}

type ListenerImpl struct {
	ctx          context.Context
	Cache        IRedis
	ListenerConn IListener

	errChan chan error
}

type Record struct {
	Id          string `json:"id"`
	CreatedDate string `json:"created_date"`
}

func New(cfg config.Config, ctx context.Context) Listener {

	lstnr := &ListenerImpl{
		ctx:     ctx,
		errChan: make(chan error),
	}
	err := lstnr.InitializeCache(cfg)
	if err != nil {
		panic(err)
	}
	lstnr.InitializeListener(cfg)

	return lstnr
}

func (l *ListenerImpl) InitializeCache(cfg config.Config) error {

	connstr := cfg.GetCacheConnDetails()

	rdb := redis.NewClient(&redis.Options{
		Addr:     connstr.Addr,
		Password: connstr.Password,
		DB:       0,
	})

	_, err := rdb.Ping(l.ctx).Result()

	if err != nil {
		return fmt.Errorf("we wanted to PONG, but instead we %s'd", err.Error())
	}

	fmt.Println("connected to cache")

	l.Cache = NewRedisRepository(rdb)
	return nil
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

func (l *ListenerImpl) Start(wait time.Duration) error {
	err := l.ListenerConn.Listen("row_inserted")
	if err != nil {
		fmt.Println("ListenerConn.Listen failed")
		return err
	}

	fmt.Println("Start monitoring PostgreSQL...")

	for {
		l.waitForNotification(l.ListenerConn, wait)
	}

}

func (listener *ListenerImpl) sendErr(err error) {
	listener.errChan <- err
}

func (listener *ListenerImpl) waitForNotification(l IListener, wait time.Duration) {
	select {
	case n := <-l.NotificationChannel():
		fmt.Println("received:", n.Extra)
		var prettyJSON bytes.Buffer
		err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")

		if err != nil {
			go listener.sendErr(fmt.Errorf("error processing JSON: %s", err.Error()))
		}

		var record Record
		json.Unmarshal([]byte(string(n.Extra)), &record)

		listener.Cache.Set(listener.ctx, record.Id, record.CreatedDate, ttl).Result()
		fmt.Printf("\nLoading into cache\n\n%s", prettyJSON.String())

	case <-time.After(wait * time.Second):
		go l.Ping()
		fmt.Println("Received no events for 90 seconds, checking connection")
	}
}

func (l *ListenerImpl) GetCache() IRedis {
	return l.Cache
}

func (l *ListenerImpl) CloseDB() error {

	if err := l.ListenerConn.Close(); err != nil {
		return fmt.Errorf("failed to close DB: %s", err.Error())
	}

	l.ListenerConn.Unlisten("row_inserted")
	log.Println("\ndb connection closed")
	return nil

}

func (l *ListenerImpl) CloseChan() {
	close(l.errChan)
}
