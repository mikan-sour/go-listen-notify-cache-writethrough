package listener

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
)

type IListener interface {
	Close() error
	Listen(channel string) error
	NotificationChannel() <-chan *pq.Notification
	Ping() error
	Unlisten(channel string) error
	UnlistenAll() error
}

type IRedis interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Ping(ctx context.Context) *redis.StatusCmd
	Set(context.Context, string, interface{}, time.Duration) *redis.StatusCmd
}
