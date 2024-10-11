package cache

import (
	"context"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"identity-server/config"
	"time"
)

type RedisCache struct {
	cache *redis.Client
}

func (i *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return i.cache.Set(ctx, key, value, ttl).Err()
}

func (i *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	val, err := i.cache.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}
	return val, true
}

func (i *RedisCache) GetAndRemove(ctx context.Context, key string) (interface{}, bool) {
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

func (i *RedisCache) Exists(ctx context.Context, key string) bool {
	res, err := i.cache.Exists(ctx, key).Result()
	if err != nil {
		return false
	}

	return res >= 1
}

func (i *RedisCache) GetOrSet(ctx context.Context, key string, fetch func() interface{}, ttl time.Duration) (interface{}, error) {
	if value, found := i.Get(ctx, key); found {
		return value, nil
	}
	value := fetch()
	if err := i.Set(ctx, key, value, ttl); err != nil {
		return nil, err
	}
	return value, nil
}

func (i *RedisCache) Remove(ctx context.Context, key string) error {
	return i.cache.Del(ctx, key).Err()
}

func NewRedisCache(config *config.RedisConfig) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Url,
		Username: config.Username,
		Password: config.Password,
		DB:       0, // use default DB
	})
	if err := redisotel.InstrumentTracing(client); err != nil {
		panic(err)
	}
	if err := redisotel.InstrumentMetrics(client); err != nil {
		panic(err)
	}

	return &RedisCache{cache: client}
}
