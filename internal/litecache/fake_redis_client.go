package litecache

import (
	"context"
	"time"

	"github.com/brave/go-sync/cache"
)

type FakeRedisClient struct {
	cache.RedisClient
	items map[string]string
}

func (c *FakeRedisClient) Set(ctx context.Context, key string, val string, ttl time.Duration) error {
	c.items[key] = val
	return nil
}

func (c *FakeRedisClient) Get(ctx context.Context, key string) (string, error) {
	value, ok := c.items[key]
	if ok {
		return value, nil
	} else {
		return "", nil
	}
}

func (c *FakeRedisClient) Del(ctx context.Context, keys ...string) error {
	for _, k := range keys {
		delete(c.items, k)
	}
	return nil
}

func (c *FakeRedisClient) FlushAll(ctx context.Context) error {
	c.items = make(map[string]string)
	return nil
}
