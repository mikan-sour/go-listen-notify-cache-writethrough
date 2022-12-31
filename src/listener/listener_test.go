package listener

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"listener_cache_writethrough/src/config"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/lib/pq"
)

var (
	cfg        config.Config
	err        error
	key, value string
)

func setup() {
	key, value = "x", "y"
	cfg, err = config.New("../../.env")
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	timeout := time.After(3 * time.Second)
	done := make(chan bool)
	go func() {
		setup()
		code := m.Run()
		done <- true
		os.Exit(code)
	}()

	select {
	case <-timeout:
		fmt.Println("Test didn't finish in time")
		os.Exit(1)
	case <-done:
	}

}

func TestE2e(t *testing.T) {
	ctx := context.Background()
	lstnr := New(cfg, context.Background())

	go lstnr.Start(1)

	dns := cfg.Dns()
	db, err := sql.Open("postgres", dns)
	if err != nil {
		panic(err)
	}

	_, err = db.QueryContext(ctx, "insert into records (created_date) values($1) returning id;", time.Now())
	if err != nil {
		t.Fatalf("expected no err but got %s", err.Error())
	}

	cache := lstnr.GetCache()

	cache.Set(ctx, key, value, 0)

	res := cache.Get(ctx, key)

	if res.Val() != value {
		t.Fatalf("expected %s but got %s", value, res.Val())
	}

}

func TestE2eWaitSeconds(t *testing.T) {
	ctx := context.Background()
	lstnr := New(cfg, context.Background())

	go lstnr.Start(1)

	dns := cfg.Dns()
	db, err := sql.Open("postgres", dns)
	if err != nil {
		panic(err)
	}

	_, err = db.QueryContext(ctx, "insert into records (created_date) values($1) returning id;", time.Now())
	if err != nil {
		t.Fatalf("expected no err but got %s", err.Error())
	}

	time.Sleep(2 * time.Second)

	cache := lstnr.GetCache()

	cache.Set(ctx, key, value, 0)

	res := cache.Get(ctx, key)

	if res.Val() != value {
		t.Fatalf("expected %s but got %s", value, res.Val())
	}

}

func TestDBCloseErrorCase(t *testing.T) {
	l := ListenerImpl{
		ctx: context.Background(),
	}
	db, _ := redismock.NewClientMock()

	l.Cache = &iredis{Client: db}

	l.ListenerConn = MockListenerImpl{
		MockClose: func() error {
			return fmt.Errorf("some error")
		},
	}

	err := l.CloseDB()
	if err == nil {
		t.Fatalf("expected %s but got nil", err.Error())
	}

}

func TestDBCloseSuccessCase(t *testing.T) {
	l := ListenerImpl{
		ctx: context.Background(),
	}
	db, mock := redismock.NewClientMock()
	mock.ExpectPing()

	l.Cache = &iredis{Client: db}

	l.ListenerConn = MockListenerImpl{
		MockClose:    func() error { return nil },
		MockUnlisten: func(channel string) error { return nil },
	}

	err := l.CloseDB()
	if err != nil {
		t.Fatalf("expected nil but got %s", err.Error())
	}

}

func TestStartFailsCase(t *testing.T) {
	l := ListenerImpl{}
	l.ListenerConn = MockListenerImpl{
		MockListen: func(channel string) error {
			return fmt.Errorf("some error")
		},
	}

	err := l.Start(1)
	if err == nil {
		t.Fatalf("expected %s but got nil", err.Error())
	}

}

func TestWaitForNotificationBadJsonCase(t *testing.T) {
	l := ListenerImpl{
		errChan: make(chan error, 1),
	}

	notifChan := make(chan *pq.Notification)
	go func() {
		notifChan <- &pq.Notification{Extra: "some bad json", Channel: "row_inserted", BePid: 1}
	}()

	db, _ := redismock.NewClientMock()

	l.Cache = &iredis{Client: db}

	l.ListenerConn = MockListenerImpl{
		MockListen: func(channel string) error {
			return nil
		},
		MockNotificationChannel: func() <-chan *pq.Notification {
			return notifChan
		},
	}

	go l.waitForNotification(l.ListenerConn, 5)

	err := <-l.errChan
	if err == nil {
		t.Fatalf("expected err but got nil")
	}

	close(notifChan)
	l.CloseChan()

}

func TestWaitForNotificationGoodJsonCase(t *testing.T) {
	called := false
	waitGroup := &sync.WaitGroup{}
	l := ListenerImpl{
		errChan: make(chan error, 1),
	}

	notifChan := make(chan *pq.Notification)
	go func() {
		notifChan <- &pq.Notification{Extra: "{\"id\":\"1\",\"created_date\":\"abc\"}", Channel: "row_inserted", BePid: 1}
	}()

	l.Cache = &iredis{Client: MockiredisImpl{
		MockPing: func(ctx context.Context) *redis.StatusCmd {
			return &redis.StatusCmd{}
		},
		MockSet: func(ctx context.Context, key string, value interface{}, exp time.Duration) *redis.StatusCmd {
			called = true
			waitGroup.Done()
			return &redis.StatusCmd{}
		},
	}}

	l.ListenerConn = MockListenerImpl{
		MockListen: func(channel string) error {
			return nil
		},
		MockNotificationChannel: func() <-chan *pq.Notification {
			return notifChan
		},
	}

	waitGroup.Add(1)
	go l.waitForNotification(l.ListenerConn, 5)
	waitGroup.Wait()

	if !called {
		t.Fatalf("expected `called` to be true but was false")
	}

}
