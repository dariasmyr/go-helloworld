package list

import (
	"errors"
	"sync"
)

type Queue[T interface{}] struct {
	ch chan T
	mu sync.Mutex
}

func NewQueue[T interface{}](capacity int) *Queue[T] {
	return &Queue[T]{
		ch: make(chan T, capacity),
	}
}

func (q *Queue[T]) Add(val T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.ch <- val
}

func (q *Queue[T]) Pop() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	select {
	case val := <-q.ch:
		return val, nil
	default:
		var zeroVal T
		return zeroVal, errors.New("queue is empty")
	}
}

func (q *Queue[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.ch)
}

func (q *Queue[T]) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.ch) == 0
}
