package data_races

import "sync"

type LimitsCache struct {
	mu     sync.RWMutex
	limits map[string]int
}

func NewLimitsCache() *LimitsCache {
	return &LimitsCache{
		limits: make(map[string]int),
	}
}

func (c *LimitsCache) GetLimit(key string) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	limit, ok := c.limits[key]
	return limit, ok
}

func (c *LimitsCache) SetLimit(key string, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.limits[key] = value
}
