package cache_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"identity-server/internal/cache"
)

func TestInMemoryCache_SetAndGet(t *testing.T) {
	c := cache.NewInMemory()

	// Test setting and getting a key
	key := "test_key"
	value := "test_value"
	c.Set(key, value, 5*time.Second)

	result, found := c.Get(key)
	assert.True(t, found, "Key should exist in cache")
	assert.Equal(t, value, result, "Returned value should match the set value")
}

func TestInMemoryCache_Get_KeyNotFound(t *testing.T) {
	c := cache.NewInMemory()

	// Test getting a non-existing key
	key := "non_existing_key"
	_, found := c.Get(key)
	assert.False(t, found, "Key should not exist in cache")
}

func TestInMemoryCache_GetOrSet(t *testing.T) {
	c := cache.NewInMemory()

	key := "fetch_key"
	value := "fetched_value"

	// Fetch function to return the value
	fetch := func() interface{} {
		return value
	}

	// Test GetOrSet when key does not exist
	result := c.GetOrSet(key, fetch, 5*time.Second)
	assert.Equal(t, value, result, "Fetched value should match the expected result")

	// Test GetOrSet when key already exists
	result = c.GetOrSet(key, fetch, 5*time.Second)
	assert.Equal(t, value, result, "Value should be retrieved from cache")
}

func TestInMemoryCache_Remove(t *testing.T) {
	c := cache.NewInMemory()

	key := "remove_key"
	value := "remove_value"

	// Set a key, ensure it's added, then remove it
	c.Set(key, value, 5*time.Second)

	result, found := c.Get(key)
	assert.True(t, found, "Key should exist in cache")
	assert.Equal(t, value, result, "Returned value should match the set value")

	// Remove the key and check if it still exists
	c.Remove(key)

	_, found = c.Get(key)
	assert.False(t, found, "Key should be removed from cache")
}

func TestInMemoryCache_TTLExpiration(t *testing.T) {
	c := cache.NewInMemory()

	key := "ttl_key"
	value := "ttl_value"

	// Set a key with a short TTL
	c.Set(key, value, 500*time.Millisecond)

	result, found := c.Get(key)
	assert.True(t, found, "Key should exist in cache")
	assert.Equal(t, value, result, "Returned value should match the set value")

	// Wait for TTL to expire
	time.Sleep(600 * time.Millisecond)

	_, found = c.Get(key)
	assert.False(t, found, "Key should have expired and not exist in cache")
}
