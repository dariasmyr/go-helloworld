package data_races

import (
	"fmt"
	"time"
)

type CacheOperation struct {
	key    string
	value  int
	result chan Result
}

type Result struct {
	Key         string
	Value       int
	Err         error
	Description CacheOperation
}

type LimitsCacheCh struct {
	data    map[string]int
	readCh  chan CacheOperation
	writeCh chan CacheOperation
}

func NewLimitsCacheCh() *LimitsCacheCh {
	cache := &LimitsCacheCh{
		data:    make(map[string]int),
		readCh:  make(chan CacheOperation),
		writeCh: make(chan CacheOperation),
	}

	go cache.processOperation()

	return cache
}

func (cache *LimitsCacheCh) processOperation() {
	for {
		select {
		case readOp := <-cache.readCh:
			fmt.Printf("[processOperation] Received read operation: key=%s\n", readOp.key)
			time.Sleep(100 * time.Millisecond)
			value, ok := cache.data[readOp.key]
			if !ok {
				fmt.Printf("[processOperation] Key not found: key=%s\n", readOp.key)
				readOp.result <- Result{Err: fmt.Errorf("key not found")}
			} else {
				fmt.Printf("[processOperation] Read successful: key=%s, value=%d\n", readOp.key, value)
				readOp.result <- Result{Key: readOp.key, Value: value}
			}
		case writeOp := <-cache.writeCh:
			fmt.Printf("[processOperation] Received write operation: key=%s, value=%d\n", writeOp.key, writeOp.value)
			time.Sleep(100 * time.Millisecond)
			cache.data[writeOp.key] = writeOp.value
			fmt.Printf("[processOperation] Write successful: key=%s, value=%d\n", writeOp.key, writeOp.value)
			writeOp.result <- Result{Key: writeOp.key, Value: writeOp.value}
		}
	}
}

func (cache *LimitsCacheCh) SetLimit(key string, value int) {
	fmt.Printf("[SetLimit] Sending write operation: key=%s, value=%d\n", key, value)
	resultCh := make(chan Result)
	cache.writeCh <- CacheOperation{key: key, value: value, result: resultCh}
	result := <-resultCh
	fmt.Printf("[SetLimit] Write operation completed: key=%s, value=%d\n", result.Key, result.Value)
}

func (cache *LimitsCacheCh) GetLimit(key string) (int, error) {
	fmt.Printf("[GetLimit] Sending read operation: key=%s\n", key)
	resultCh := make(chan Result)
	cache.readCh <- CacheOperation{key: key, result: resultCh}
	result := <-resultCh
	if result.Err != nil {
		fmt.Printf("[GetLimit] Read operation failed: key=%s, error=%v\n", key, result.Err)
		return 0, result.Err
	}
	fmt.Printf("[GetLimit] Read operation completed: key=%s, value=%d\n", result.Key, result.Value)
	return result.Value, nil
}
