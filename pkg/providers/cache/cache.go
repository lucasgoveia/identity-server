package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (interface{}, bool)
	GetOrSet(ctx context.Context, key string, fetch func() interface{}, ttl time.Duration) (interface{}, error)
	Remove(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) bool
	GetAndRemove(ctx context.Context, key string) (interface{}, bool)
}
