package data_races

import (
	"fmt"
	"testing"
)

func TestAsyncBehaviorPool(t *testing.T) {
	cache := NewLimitsCachePool(2)

	go cache.Run()

	for i := 0; i < 10; i++ {
		sendReq := Request{
			Op:    "set",
			Key:   fmt.Sprintf("key%d", i),
			Value: i * 10,
		}
		fmt.Printf("Sending SENT req for %v: %d\n", sendReq.Key, sendReq.Value)
		cache.requests <- sendReq
	}

	for i := 0; i < 10; i++ {
		getReq := Request{
			Op:  "get",
			Key: fmt.Sprintf("key%d", i),
		}
		fmt.Printf("Sending GET req for %v\n", getReq.Key)
		cache.requests <- getReq

	}

	fmt.Printf("closing req channel")
	close(cache.requests)

	results := make([]Response, 20)
	i := 0
	for result := range cache.results {
		results[i] = result
		i++
	}

	for _, result := range results {
		if result.Ok != true {
			t.Errorf("expected no error, got %v", result.Limit)
		}
		if result.Limit == 0 {
			t.Errorf("expected a value, got nil")
		}
	}
}
