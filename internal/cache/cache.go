package cache

import "time"

type Cache interface {
	Set(key string, value interface{}, ttl time.Duration)
	Get(key string) (interface{}, bool)
	GetOrSet(key string, fetch func() interface{}, ttl time.Duration) interface{}
	Remove(key string)
}
