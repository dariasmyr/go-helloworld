package queue

import (
	"sync"
	"testing"
)

func TestQueueBasicOperations(t *testing.T) {
	q := NewQueue[int]()

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
}

func TestQueueHighLoadAdd(t *testing.T) {
	q := NewQueue[int]()

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
}

func TestQueueHighLoadAddAndPio(t *testing.T) {
	q := NewQueue[int]()

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
}
