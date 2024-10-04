package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"identity-server/config"
	"time"
)

type RedisCache struct {
	cache *redis.Client
}

func (i *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	return i.cache.Set(context.Background(), key, value, ttl).Err()
}

func (i *RedisCache) Get(key string) (interface{}, bool) {
	val, err := i.cache.Get(context.Background(), key).Result()
	if err != nil {
		return nil, false
	}
	return val, true
}

func (i *RedisCache) Exists(key string) bool {
	res, err := i.cache.Exists(context.Background(), key).Result()
	if err != nil {
		return false
	}

	return res >= 1
}

func (i *RedisCache) GetOrSet(key string, fetch func() interface{}, ttl time.Duration) (interface{}, error) {
	if value, found := i.Get(key); found {
		return value, nil
	}
	value := fetch()
	if err := i.Set(key, value, ttl); err != nil {
		return nil, err
	}
	return value, nil
}

func (i *RedisCache) Remove(key string) error {
	return i.cache.Del(context.Background(), key).Err()
}

func NewRedisCache(config *config.RedisConfig) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Url,
		Username: config.Username,
		Password: config.Password, // no password set
		DB:       0,               // use default DB
	})
	return &RedisCache{cache: client}
}
