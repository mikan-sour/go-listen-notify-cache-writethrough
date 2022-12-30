package listener

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"listener_cache_writethrough/src/config"
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
	setup()
	code := m.Run()
	os.Exit(code)
}

func TestE2e(t *testing.T) {
	ctx := context.Background()
	lstnr := New(cfg, context.Background())

	go lstnr.Start()

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
