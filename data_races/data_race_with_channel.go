package data_races

import (
	"fmt"
)

type CacheOperationAsync struct {
	key           string
	value         int
	operationType string
	result        chan Result
}

type Result struct {
	Key         string
	Value       int
	Err         error
	Description CacheOperationAsync
}

type LimitsCacheCh struct {
	data   map[string]int
	taskCh chan CacheOperationAsync
}

func NewLimitsCacheCh() *LimitsCacheCh {
	cache := &LimitsCacheCh{
		data:   make(map[string]int),
		taskCh: make(chan CacheOperationAsync),
	}

	go cache.processOperation()

	return cache
}

func (cache *LimitsCacheCh) processOperation() {
	for op := range cache.taskCh {
		switch op.operationType {
		case "set":
			cache.data[op.key] = op.value
			op.result <- Result{Key: op.key, Value: cache.data[op.key]}
		case "get":
			value, ok := cache.data[op.key]
			if !ok {
				op.result <- Result{Err: fmt.Errorf("key not found")}
			} else {
				op.result <- Result{Key: op.key, Value: value}
			}
		}
	}
}

func (cache *LimitsCacheCh) SetLimit(key string, value int) {
	fmt.Printf("[SetLimit] Sending write operation: key=%s, value=%d\n", key, value)
	resultCh := make(chan Result)
	cache.taskCh <- CacheOperationAsync{key: key, value: value, result: resultCh, operationType: "set"}
	result := <-resultCh
	fmt.Printf("[SetLimit] Write operation completed: key=%s, value=%d\n", result.Key, result.Value)
}

func (cache *LimitsCacheCh) GetLimit(key string) (int, error) {
	fmt.Printf("[GetLimit] Sending read operation: key=%s\n", key)
	resultCh := make(chan Result)
	cache.taskCh <- CacheOperationAsync{key: key, result: resultCh, operationType: "get"}
	result := <-resultCh
	if result.Err != nil {
		fmt.Printf("[GetLimit] Read operation failed: key=%s, error=%v\n", key, result.Err)
		return 0, result.Err
	}
	fmt.Printf("[GetLimit] Read operation completed: key=%s, value=%d\n", result.Key, result.Value)
	return result.Value, nil
}
