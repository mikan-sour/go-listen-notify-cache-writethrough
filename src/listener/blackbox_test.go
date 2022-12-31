package listener

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"
)

func skipT(t *testing.T) {
	if os.Getenv("E2E") != "1" {
		t.Skip("Skipping testing in CI environment")
	}
}

func TestE2e(t *testing.T) {
	skipT(t)

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
	skipT(t)

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
