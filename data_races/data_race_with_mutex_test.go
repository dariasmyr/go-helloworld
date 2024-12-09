package data_races

import (
	"sync"
	"testing"
)

func TestLimitsCache_GetSet(t *testing.T) {
	cache := NewLimitsCache()

	cache.SetLimit("key1", 100)
	if limit, ok := cache.GetLimit("key1"); !ok || limit != 100 {
		t.Errorf("expected limit 100, got %d", limit)
	}
}

func TestLimitsCache_ConcurrentAccess(t *testing.T) {
	cache := NewLimitsCache()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.SetLimit(string(rune('A'+i%26)), i)
		}(i)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.GetLimit(string(rune('A' + i%26)))
		}(i)
	}

	wg.Wait()
}

func TestLimitsCache_NoRaceConditions(t *testing.T) {
	cache := NewLimitsCache()
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := string(rune('A' + i%26))
			cache.SetLimit(key, i)
			cache.GetLimit(key)
		}(i)
	}

	wg.Wait()
}
