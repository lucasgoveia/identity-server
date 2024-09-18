package cache

type Cache interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	GetOrSet(key string, fetch func() interface{}) interface{}
	GetWithTTL(key string) (interface{}, bool, int64)
	Remove(key string)
}
