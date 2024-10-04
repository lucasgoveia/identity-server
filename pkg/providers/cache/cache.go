package cache

import "time"

type Cache interface {
	Set(key string, value interface{}, ttl time.Duration) error
	Get(key string) (interface{}, bool)
	GetOrSet(key string, fetch func() interface{}, ttl time.Duration) (interface{}, error)
	Remove(key string) error
	Exists(key string) bool
}
