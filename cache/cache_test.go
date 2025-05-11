package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type cacheVal struct {
	val       interface{}
	expiresAt time.Time
}

type Cache struct {
	mu      sync.Mutex
	cache   map[string]*cacheVal
	ttl     time.Duration
	context context.Context
	cancel  context.CancelFunc
}

func NewCache(ttl time.Duration) *Cache {
	ctx, cancel := context.WithCancel(context.Background())
	cache := &Cache{
		cache:   make(map[string]*cacheVal),
		ttl:     ttl,
		context: ctx,
		cancel:  cancel,
	}

	go cache.cleanup()

	return cache
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(c.ttl)

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for k, v := range c.cache {
				if time.Now().After(v.expiresAt) {
					delete(c.cache, k)
				}
			}
			c.mu.Unlock()
		case <-c.context.Done():
			fmt.Println("context done")
			return
		}
	}
}

func (c *Cache) Set(k string, v interface{}) {
	c.mu.Lock()
	c.cache[k] = &cacheVal{
		val:       v,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

func (c *Cache) Get(k string) (interface{}, bool) {
	c.mu.Lock()
	if v, ok := c.cache[k]; ok {
		c.mu.Unlock()
		return v.val, true
	}
	c.mu.Unlock()
	return nil, false
}

func (c *Cache) Delete(k string) {
	c.mu.Lock()
	delete(c.cache, k)
	c.mu.Unlock()
}

func (c *Cache) Close() {
	c.cancel()
}

func TestCache(t *testing.T) {
	cache := NewCache(10 * time.Second)
	for i := 0; i <= 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		fmt.Println("Setting key", key)
		cache.Set(key, "Value")
		if i%2 == 0 {
			fmt.Println("Delete odd key", key)
			cache.Delete(key)
		}
	}

	for i := 0; i <= 10; i++ {
		key := fmt.Sprintf("key-%d", i)
		if val, ok := cache.Get(key); ok {
			if strVal, valid := val.(string); !valid {
				fmt.Printf("Value type is invalid for key %s\n", key)
			} else {
				fmt.Printf("Key %s = %s \n", key, strVal)
			}
		} else {
			fmt.Printf("Key %s expired of not found \n", key)
		}
	}

	cache.Close()
	time.Sleep(1 * time.Second)
}
