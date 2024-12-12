package data_races

import (
	"fmt"
	"testing"
)

func TestAsyncBehavior(t *testing.T) {
	cache := NewLimitsCachePool()

	requests := make(chan Request, 100)

	go cache.HandleRequests(requests)

	for i := 0; i < 10; i++ {
		respChan := make(chan Response)
		sendReq := Request{
			Op:    "set",
			Key:   fmt.Sprintf("key%d", i),
			Value: i * 10,
			Resp:  respChan,
		}
		fmt.Printf("Sending SENT req for key%v: %d\n", sendReq.Key, sendReq.Value)
		requests <- sendReq

		resp := <-respChan
		if resp.Ok {
			fmt.Printf("Set limit for key%d: %d\n", i, i*10)
		}
	}

	for i := 0; i < 10; i++ {
		respChan := make(chan Response)
		getReq := Request{
			Op:   "get",
			Key:  fmt.Sprintf("key%d", i),
			Resp: respChan,
		}
		fmt.Printf("Sending GET req for key%v\n", getReq.Key)
		requests <- getReq
		resp := <-respChan
		if resp.Ok {
			fmt.Printf("Limit for key%d: %d\n", i, resp.Limit)
		} else {
			fmt.Printf("No limit for key%d\n", i)
		}
	}

	fmt.Printf("closing req channel")

	close(requests)
}
