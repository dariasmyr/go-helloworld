package queue

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	t.Run("Test Simple Queue Channel", func(t *testing.T) {
		err := runQueueChannel(100)
		if err != nil {
			return
		}
	})

	t.Run("Test Simple Queue List", func(t *testing.T) {
		err := runQueueList(100)
		if err != nil {
			return
		}
	})
}

func runQueueChannel(cap int) error {
	queue := make(chan int, 100)

	go func() {
		for i := 0; i < cap; i++ {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("Writing", i)
			queue <- i
		}
		close(queue)
	}()

	for {
		select {
		case i, ok := <-queue:
			if !ok {
				fmt.Println("Queue is empty (channel closed)")
				return nil
			}
			fmt.Println("Received ", i)
		default:
			fmt.Println("Queue is empty (no more elements to receive), waiting...")
			time.Sleep(100 * time.Millisecond)
		}
	}
}

type Node struct {
	value int
	next  *Node
}

type SimpleQueue struct {
	mu    sync.Mutex
	Last  *Node
	First *Node
}

func (q *SimpleQueue) Push(value int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	n := &Node{value: value}

	if q.First == nil {
		q.First = n
		q.Last = n
	} else {
		q.Last.next = n
		q.Last = n
	}
}

func (q *SimpleQueue) Pop() (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.First == nil {
		return 0, fmt.Errorf("queue list is empty")
	}

	result := q.First.value
	q.First = q.First.next

	return result, nil
}

func (q *SimpleQueue) Peek() (int, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.First == nil {
		return 0, fmt.Errorf("queue list is empty")
	}

	return q.First.value, nil
}

func runQueueList(cap int) error {
	queue := SimpleQueue{
		mu: sync.Mutex{},
	}

	for i := 0; i < cap; i++ {
		queue.Push(i)
	}

	peekedValue, err := queue.Peek()
	if err != nil {
		return err
	}
	fmt.Println("Peek queue", peekedValue)

	for i := 0; i < cap; i++ {
		result, err := queue.Pop()
		if err != nil {
			return err
		}
		fmt.Println(result)
	}

	return nil
}
