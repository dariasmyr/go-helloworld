package cache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type entry struct {
	key       string
	val       interface{}
	expiresAt time.Time
}

type LRUCache struct {
	mu       sync.Mutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
	ttl      time.Duration
	context  context.Context
	cancel   context.CancelFunc
}

func NewLRUCache(cap int, ttl time.Duration) *LRUCache {
	ctx, cancel := context.WithCancel(context.Background())
	cache := &LRUCache{
		capacity: cap,
		items:    make(map[string]*list.Element),
		order:    list.New(),
		ttl:      ttl,
		context:  ctx,
		cancel:   cancel,
	}

	return cache
}

func (c *LRUCache) Set(k string, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[k]; exists {
		c.order.MoveToFront(elem)
		elem.Value.(*entry).val = v
		elem.Value.(*entry).expiresAt = time.Now().Add(c.ttl)
		return
	}

	if c.order.Len() >= c.capacity {
		c.evictOldest()
	}

	ent := &entry{k, v, time.Now().Add(c.ttl)}
	elem := c.order.PushFront(ent)
	c.items[k] = elem
}

func (c *LRUCache) Get(k string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.items[k]
	if !ok {
		return nil, false
	}

	ent := elem.Value.(*entry)
	if time.Now().After(ent.expiresAt) {
		c.removeElement(elem)
		return nil, false
	}

	c.order.MoveToFront(elem)
	return ent.val, true
}

func (c *LRUCache) Delete(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[k]; ok {
		c.removeElement(elem)
	}
}

func (c *LRUCache) Close() {
	c.cancel()
}

func (c *LRUCache) evictOldest() {
	elem := c.order.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *LRUCache) removeElement(elem *list.Element) {
	ent := elem.Value.(*entry)
	delete(c.items, ent.key)
	c.order.Remove(elem)
}

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(5, 5*time.Second)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)
	cache.Set("d", 4)
	cache.Set("e", 5)

	cache.Get("a") // Use a, so a becomes recently used

	cache.Set("f", 6) // b will be deleted as most not used

	if val, ok := cache.Get("b"); ok {
		fmt.Println("b =", val)
	} else {
		fmt.Println("b not found (evicted or expired)")
	}

	cache.Close()
}
