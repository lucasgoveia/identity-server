package cache

import (
	"context"
	"github.com/dgraph-io/ristretto"
	"time"
)

type InMemoryCache struct {
	cache *ristretto.Cache[string, any]
}

func (i *InMemoryCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	i.cache.SetWithTTL(key, value, 1, ttl)
	i.cache.Wait()
	return nil
}

func (i *InMemoryCache) Get(_ context.Context, key string) (interface{}, bool) {
	return i.cache.Get(key)
}

func (i *InMemoryCache) GetAndRemove(ctx context.Context, key string) (interface{}, bool) {
	val, exists := i.Get(ctx, key)

	if !exists {
		return nil, false
	}

	err := i.Remove(ctx, key)

	if err != nil {
		return nil, false
	}

	return val, true
}

func (i *InMemoryCache) Exists(_ context.Context, key string) bool {
	_, found := i.cache.Get(key)
	return found
}

func (i *InMemoryCache) GetOrSet(ctx context.Context, key string, fetch func() interface{}, ttl time.Duration) (interface{}, error) {
	if value, found := i.Get(ctx, key); found {
		return value, nil
	}
	value := fetch()
	if err := i.Set(ctx, key, value, ttl); err != nil {
		return nil, err
	}
	return value, nil
}

func (i *InMemoryCache) Remove(_ context.Context, key string) error {
	i.cache.Del(key)
	return nil
}

func NewInMemory() *InMemoryCache {
	cache, err := ristretto.NewCache[string, any](&ristretto.Config[string, any]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
	return &InMemoryCache{cache: cache}
}
