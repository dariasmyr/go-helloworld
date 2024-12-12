package data_races

import (
	"fmt"
	"sync"
)

type LimitsCachePool struct {
	numWorkers int
	requests   chan Request
	results    chan Response
	limits     map[string]int
}

func NewLimitsCachePool(nw int) *LimitsCachePool {

	return &LimitsCachePool{
		numWorkers: nw,
		limits:     make(map[string]int),
		requests:   make(chan Request, nw),
		results:    make(chan Response, nw),
	}

}

func (c *LimitsCachePool) Run() {
	var wg sync.WaitGroup

	for i := 0; i < c.numWorkers; i++ {
		wg.Add(1)
		go c.HandleRequests(&wg)
	}

	wg.Wait()
	close(c.results)
}

func (c *LimitsCachePool) HandleRequests(wg *sync.WaitGroup) {
	defer wg.Done()
	for req := range c.requests {
		switch req.Op {
		case "get":
			if limit, ok := c.limits[req.Key]; ok {
				c.results <- Response{Limit: limit, Ok: true}
				fmt.Printf("Limit for key%v: %d\n", req.Key, limit)
			} else {
				c.results <- Response{Ok: false}
				fmt.Printf("No limit for key%v\n", req.Key)
			}
		case "set":
			c.limits[req.Key] = req.Value
			c.results <- Response{Ok: true}
			fmt.Printf("Set limit for key%v: %d\n", req.Key, req.Value)
		}
	}
}
