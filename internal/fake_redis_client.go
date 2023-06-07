package internal

import (
	"context"
	"time"

	"github.com/brave/go-sync/cache"
	lru "github.com/hashicorp/golang-lru"
)

const cacheSize = 1024

type FakeRedisClient struct {
	cache.RedisClient
	items *lru.Cache
}

func NewFakeRedisClient() *FakeRedisClient {
	cache, _ := lru.New(cacheSize)
	return &FakeRedisClient{items: cache}
}

func (c *FakeRedisClient) Set(ctx context.Context, key string, val string, ttl time.Duration) error {
	c.items.Add(key, val)
	return nil
}

func (c *FakeRedisClient) Get(ctx context.Context, key string) (string, error) {
	value, ok := c.items.Get(key)
	if ok {
		return value.(string), nil
	} else {
		return "", nil
	}
}

func (c *FakeRedisClient) Del(ctx context.Context, keys ...string) error {
	for _, k := range keys {
		c.items.Remove(k)
	}
	return nil
}

func (c *FakeRedisClient) FlushAll(ctx context.Context) error {
	c.items, _ = lru.New(cacheSize)
	return nil
}
