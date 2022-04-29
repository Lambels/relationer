package redis

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type CacheService struct {
	cache *cache.Cache
}

func NewCache(addr string) *CacheService {
	c := CacheService{}

	rdb := redis.NewClient(
		&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		},
	)
	cc := cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	c.cache = cc
	return &c
}

func (c *CacheService) Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	return c.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: val,
		TTL:   ttl,
	})
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
	return c.cache.Delete(ctx, key)
}

func (c *CacheService) Get(ctx context.Context, key string, val interface{}) error {
	return c.cache.Get(ctx, key, val)
}
