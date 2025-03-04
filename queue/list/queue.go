package list

import (
	"errors"
	"sync"
)

type Queue[T interface{}] struct {
	first *node[T]
	last  *node[T]
	count int
	mu    sync.Mutex
}

type node[T interface{}] struct {
	next  *node[T]
	value T
}

func NewQueue[T interface{}]() *Queue[T] {
	return &Queue[T]{}
}

func (q *Queue[T]) Add(val T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	n := &node[T]{value: val}

	if q.count == 0 {
		q.first, q.last = n, n
	} else {
		q.last.next = n
		q.last = n
	}
	q.count++
}

func (q *Queue[T]) Pop() (T, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.count == 0 {
		var zeroVal T
		return zeroVal, errors.New("queue is empty")
	}

	val := q.first.value
	q.first = q.first.next
	q.count--

	if q.count == 0 {
		q.last = nil
	}

	return val, nil
}

func (q *Queue[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.count
}

func (q *Queue[T]) IsEmpty() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.count == 0
}
