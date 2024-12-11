package data_races

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestAsyncBehavior(t *testing.T) {
	cache := NewLimitsCacheCh()

	start := time.Now()
	fmt.Println("[TestAsyncBehavior] Starting test")
	go cache.SetLimit("key1", 100)
	go cache.SetLimit("key2", 200)
	go cache.GetLimit("key1")
	go cache.GetLimit("key2")

	time.Sleep(150 * time.Millisecond)
	duration := time.Since(start)
	if duration > 500*time.Millisecond {
		t.Errorf("[TestAsyncBehavior] Operations are not asynchronous, took too long: %v", duration)
	}
	fmt.Println("[TestAsyncBehavior] Test completed")
}

func TestConcurrentAsync(t *testing.T) {
	cache := NewLimitsCacheCh()
	const numOps = 100
	fmt.Println("[TestConcurrentAsync] Starting test")
	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < numOps; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			fmt.Printf("[TestConcurrentAsync] Starting operation: key=%s, value=%d\n", key, i)
			cache.SetLimit(key, i)
			_, _ = cache.GetLimit(key)
			fmt.Printf("[TestConcurrentAsync] Completed operation: key=%s\n", key)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	fmt.Printf("[TestConcurrentAsync] Completed %d operations in %v\n", numOps, duration)
}
