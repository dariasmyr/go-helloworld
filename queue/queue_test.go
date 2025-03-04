package queue

import (
	channel "go-helloworld/queue/channel"
	list "go-helloworld/queue/list"
	"runtime"
	"sync"
	"testing"
)

type Queue[T any] interface {
	Add(val T)
	Pop() (T, error)
	Size() int
	IsEmpty() bool
}

func TestQueueBasicOperations(t *testing.T) {
	tests := []struct {
		name string
		q    Queue[int]
	}{
		{"List Queue", list.NewQueue[int]()},
		{"Channel Queue", channel.NewQueue[int](
			1000000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.q

			q.Add(1)
			q.Add(2)
			q.Add(3)

			if q.Size() != 3 {
				t.Errorf("Expected queue size to be 3, but got %d", q.Size())
			}

			val, _ := q.Pop()
			if val != 1 {
				t.Errorf("Expected 1, but got %d", val)
			}

			val, _ = q.Pop()
			if val != 2 {
				t.Errorf("Expected 2, but got %d", val)
			}

			val, _ = q.Pop()
			if val != 3 {
				t.Errorf("Expected 3, but got %d", val)
			}

			if !q.IsEmpty() {
				t.Errorf("Expected queue to be empty, but it's not")
			}
		})
	}
}

func TestQueueHighLoadAdd(t *testing.T) {
	tests := []struct {
		name string
		q    Queue[int]
	}{
		{"List Queue", list.NewQueue[int]()},
		{"Channel Queue", channel.NewQueue[int](
			1000000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.q

			var wg sync.WaitGroup
			numTasks := 1000000

			for i := 0; i < numTasks; i++ {
				wg.Add(1)
				go func(i int) {
					q.Add(i)
					wg.Done()
				}(i)
			}

			wg.Wait()
			if q.Size() != numTasks {
				t.Errorf("Expected queue size to be 1000000, but got %d", q.Size())
			}
		})
	}
}

func TestQueueHighLoadAddAndPop(t *testing.T) {
	tests := []struct {
		name string
		q    Queue[int]
	}{
		{"List Queue", list.NewQueue[int]()},
		{"Channel Queue", channel.NewQueue[int](1000000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.q

			var wg sync.WaitGroup
			numTasks := 1000000

			for i := 0; i < numTasks; i++ {
				wg.Add(1)
				go func(i int) {
					q.Add(i)
					wg.Done()
				}(i)
			}

			for i := 0; i < numTasks; i++ {
				wg.Add(1)
				go func() {
					q.Pop()
					wg.Done()
				}()
			}

			wg.Wait()

			if !q.IsEmpty() {
				t.Errorf("Expected queue to be empty, but it's not")
			}
		})
	}
}

func TestQueueWithLimitProcessors(t *testing.T) {
	tests := []struct {
		name string
		q    Queue[int]
	}{
		{"List Queue", list.NewQueue[int]()},
		{"Channel Queue", channel.NewQueue[int](1000000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.q
			numTasks := 1000000
			maxProcCount := 5

			for i := 0; i < numTasks; i++ {
				q.Add(i)
			}

			var wg sync.WaitGroup
			maxActiveWorkers := 10

			runtime.GOMAXPROCS(maxProcCount)

			for i := 0; i < maxActiveWorkers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						if q.IsEmpty() {
							break
						}
						q.Pop()
					}
				}()
			}

			wg.Wait()

			if !q.IsEmpty() {
				t.Errorf("Expected queue to be empty, but it's not")
			}
		})
	}
}
