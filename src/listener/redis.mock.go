package listener

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type IRedisImpl struct {
	Client redis.Cmdable
}

func NewRedisRepository(Client redis.Cmdable) IRedis {
	return &IRedisImpl{Client}
}

// Set attaches the redis repository and set the data
func (r *IRedisImpl) Set(ctx context.Context, key string, value interface{}, exp time.Duration) *redis.StatusCmd {
	return r.Client.Set(ctx, key, value, exp)
}

// Get attaches the redis repository and get the data
func (r *IRedisImpl) Get(ctx context.Context, key string) *redis.StringCmd {
	return r.Client.Get(ctx, key)
}

func (r *IRedisImpl) Ping(ctx context.Context) *redis.StatusCmd {
	return r.Client.Ping(ctx)
}

type MockIRedisImpl struct {
	redis.Cmdable
	MockSet  func(ctx context.Context, key string, value interface{}, exp time.Duration) *redis.StatusCmd
	MockGet  func(ctx context.Context, key string) *redis.StringCmd
	MockPing func(ctx context.Context) *redis.StatusCmd
}

func (m MockIRedisImpl) Set(ctx context.Context, key string, value interface{}, exp time.Duration) *redis.StatusCmd {
	return m.MockSet(ctx, key, value, exp)
}
func (m MockIRedisImpl) Get(ctx context.Context, key string) *redis.StringCmd {
	return m.MockGet(ctx, key)
}
func (m MockIRedisImpl) Ping(ctx context.Context) *redis.StatusCmd {
	return m.MockPing(ctx)
}
