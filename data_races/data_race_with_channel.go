package data_races

import "fmt"

type LimitsCacheOneGorourine struct {
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

func NewLimitsCacheOneGorourine() *LimitsCacheOneGorourine {
	return &LimitsCacheOneGorourine{
		limits: make(map[string]int),
	}
}

func (c *LimitsCacheOneGorourine) HandleRequests(requests chan Request) {
	for req := range requests {
		switch req.Op {
		case "get":
			if limit, ok := c.limits[req.Key]; ok {
				req.Resp <- Response{Limit: limit, Ok: true}
				fmt.Printf("Limit for key%v: %d\n", req.Key, limit)
			} else {
				req.Resp <- Response{Ok: false}
				fmt.Printf("No limit for key%v\n", req.Key)
			}
		case "set":
			c.limits[req.Key] = req.Value
			req.Resp <- Response{Ok: true}
			fmt.Printf("Set limit for key%v: %d\n", req.Key, req.Value)
		}
	}
}
