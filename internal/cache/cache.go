package cache

type Cache interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	GetOrSet(key string, fetch func() interface{}) interface{}
	Remove(key string)
}
