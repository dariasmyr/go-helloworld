package goroutines_safety

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// RunTasksWithChannels runs multiple tasks in parallel.
// If one of the tasks causes a panic, it is processed with recover() and passed to errChan.
// tasks - the number of tasks, task Func - a function to perform, errChan - a channel for errors, failOn - the number of a task that causes panic.
func RunTasksWithChannels(tasks int, taskFunc func(int, int) func(), errChan chan error, failOn int) {
	var wg sync.WaitGroup
	wg.Add(tasks)

	for i := 0; i < tasks; i++ {
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errChan <- fmt.Errorf("task %d panicked: %v", id, r)
				}
			}()
			taskFunc(id, failOn)()
		}(i)
	}

	wg.Wait()

	close(errChan)
}

func mockTask(id int, failOn int) func() {
	return func() {
		if id == failOn {
			panic("Task failed")
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func TestRunTasksWithChannels(t *testing.T) {
	t.Run("All tasks complete without panic", func(t *testing.T) {
		tasks := 5
		errChan := make(chan error, 1)

		go RunTasksWithChannels(tasks, mockTask, errChan, -1) // No errors (failOn=-1)

		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		case <-time.After(1 * time.Second):
		}

	})

	t.Run("Task panics and propagates error", func(t *testing.T) {
		tasks := 5
		errChan := make(chan error, 1)

		go RunTasksWithChannels(tasks, mockTask, errChan, 2) // Ошибка на задаче 2

		select {
		case err := <-errChan:
			if err == nil {
				t.Error("expected error, got nil")
			}
		case <-time.After(1 * time.Second):
			t.Error("timeout waiting for error")
		}
	})
}
