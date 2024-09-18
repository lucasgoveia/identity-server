package cache

import (
	"github.com/dgraph-io/ristretto"
)

type InMemoryCache struct {
	cache *ristretto.Cache[string, any]
}

func (i *InMemoryCache) Set(key string, value interface{}) {
	i.cache.Set(key, value, 1)
	i.cache.Wait()
}

func (i *InMemoryCache) Get(key string) (interface{}, bool) {
	return i.cache.Get(key)
}

func (i *InMemoryCache) GetOrSet(key string, fetch func() interface{}) interface{} {
	if value, found := i.Get(key); found {
		return value
	}
	value := fetch()
	i.Set(key, value)
	return value
}

func (i *InMemoryCache) Remove(key string) {
	i.cache.Del(key)
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
