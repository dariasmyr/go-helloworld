package data_races

import "fmt"

type LimitsCachePool struct {
	limits map[string]int
}

type Request struct {
	Op    string // Операция: "get" или "set"
	Key   string
	Value int
	Resp  chan Response // Канал для получения ответа
}

type Response struct {
	Limit int  // Значение лимита
	Ok    bool // Успешно ли выполнена операция
}

func NewLimitsCachePool() *LimitsCachePool {
	return &LimitsCachePool{
		limits: make(map[string]int),
	}
}

func (c *LimitsCachePool) HandleRequests(requests chan Request) {
	for req := range requests {
		switch req.Op {
		case "get":
			// Выполняем операцию получения лимита
			if limit, ok := c.limits[req.Key]; ok {
				req.Resp <- Response{Limit: limit, Ok: true}
				fmt.Printf("Limit for key%v: %d\n", req.Key, limit)
			} else {
				req.Resp <- Response{Ok: false}
				fmt.Printf("No limit for key%v\n", req.Key)
			}
		case "set":
			// Выполняем операцию установки лимита
			c.limits[req.Key] = req.Value
			req.Resp <- Response{Ok: true}
			fmt.Printf("Set limit for key%v: %d\n", req.Key, req.Value)
		}
	}
}
