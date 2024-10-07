package cache_test

import (
	"context"
	"identity-server/pkg/providers/cache"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryCache_SetAndGet(t *testing.T) {
	ctx := context.Background()
	c := cache.NewInMemory()

	// Test setting and getting a key
	key := "test_key"
	value := "test_value"
	err := c.Set(ctx, key, value, 5*time.Second)
	assert.NoError(t, err)
	result, found := c.Get(ctx, key)
	assert.True(t, found, "Key should exist in cache")
	assert.Equal(t, value, result, "Returned value should match the set value")
}

func TestInMemoryCache_Get_KeyNotFound(t *testing.T) {
	c := cache.NewInMemory()
	ctx := context.Background()

	// Test getting a non-existing key
	key := "non_existing_key"
	_, found := c.Get(ctx, key)
	assert.False(t, found, "Key should not exist in cache")
}

func TestInMemoryCache_GetOrSet(t *testing.T) {
	ctx := context.Background()
	c := cache.NewInMemory()

	key := "fetch_key"
	value := "fetched_value"

	// Fetch function to return the value
	fetch := func() interface{} {
		return value
	}

	// Test GetOrSet when key does not exist
	result, err := c.GetOrSet(ctx, key, fetch, 5*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, value, result, "Fetched value should match the expected result")

	// Test GetOrSet when key already exists
	result, err = c.GetOrSet(ctx, key, fetch, 5*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, value, result, "Value should be retrieved from cache")
}

func TestInMemoryCache_Remove(t *testing.T) {
	c := cache.NewInMemory()
	ctx := context.Background()

	key := "remove_key"
	value := "remove_value"

	// Set a key, ensure it's added, then remove it
	err := c.Set(ctx, key, value, 5*time.Second)
	assert.NoError(t, err)

	result, found := c.Get(ctx, key)
	assert.True(t, found, "Key should exist in cache")
	assert.Equal(t, value, result, "Returned value should match the set value")

	// Remove the key and check if it still exists
	err = c.Remove(ctx, key)

	assert.NoError(t, err)
	_, found = c.Get(ctx, key)
	assert.False(t, found, "Key should be removed from cache")
}

func TestInMemoryCache_TTLExpiration(t *testing.T) {
	c := cache.NewInMemory()
	ctx := context.Background()

	key := "ttl_key"
	value := "ttl_value"

	// Set a key with a short TTL
	err := c.Set(ctx, key, value, 500*time.Millisecond)

	assert.NoError(t, err)
	result, found := c.Get(ctx, key)
	assert.True(t, found, "Key should exist in cache")
	assert.Equal(t, value, result, "Returned value should match the set value")

	// Wait for TTL to expire
	time.Sleep(600 * time.Millisecond)

	_, found = c.Get(ctx, key)
	assert.False(t, found, "Key should have expired and not exist in cache")
}
