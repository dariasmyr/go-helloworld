package list

import (
	"errors"
)

type Queue[T interface{}] struct {
	ch chan T
}

func NewQueue[T interface{}](capacity int) *Queue[T] {
	return &Queue[T]{
		ch: make(chan T, capacity),
	}
}

func (q *Queue[T]) Add(val T) {
	q.ch <- val
}

func (q *Queue[T]) Pop() (T, error) {
	select {
	case val := <-q.ch:
		return val, nil
	default:
		var zeroVal T
		return zeroVal, errors.New("queue is empty")
	}
}

func (q *Queue[T]) Size() int {
	return len(q.ch)
}

func (q *Queue[T]) IsEmpty() bool {
	return len(q.ch) == 0
}
