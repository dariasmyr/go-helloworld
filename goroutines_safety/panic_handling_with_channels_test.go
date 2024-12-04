package goroutines_safety

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// RunTasksWithChannels runs multiple tasks in parallel.
// If one of the tasks causes a panic, it is processed with recover() and passed to errChan.
// tasks - the number of tasks, taskFunc - a function to perform, errChan - a channel for errors, failOn - the number of a task that causes panic.
func RunTasksWithChannels(tasks int, taskFunc func(int, int) func(), errChan chan error, failOn int) {
	fmt.Println("[RunTasksWithChannels] Starting tasks...")

	var wg sync.WaitGroup
	wg.Add(tasks)

	for i := 0; i < tasks; i++ {
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("task %d panicked: %v", id, r)
					fmt.Println("[RunTasksWithChannels] Recovered from panic:", err)
					errChan <- err
				}
			}()

			fmt.Printf("[RunTasksWithChannels] Task %d started\n", id)
			taskFunc(id, failOn)()
			fmt.Printf("[RunTasksWithChannels] Task %d completed\n", id)
		}(i)
	}

	wg.Wait()

	fmt.Println("[RunTasksWithChannels] All tasks completed. Closing error channel.")
	close(errChan)
}

func mockTask(id int, failOn int) func() {
	return func() {
		fmt.Printf("[mockTask] Task %d executing\n", id)
		if id == failOn {
			fmt.Printf("[mockTask] Task %d will panic\n", id)
			panic("Task failed")
		}
		time.Sleep(100 * time.Millisecond) // Mock task executing
		fmt.Printf("[mockTask] Task %d finished executing\n", id)
	}
}

func TestRunTasksWithChannels(t *testing.T) {
	t.Run("All tasks complete without panic", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithChannels] Starting test: All tasks complete without panic")

		tasks := 5
		errChan := make(chan error, 1)

		go RunTasksWithChannels(tasks, mockTask, errChan, -1) // No errors (failOn=-1)

		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("[TestRunTasksWithChannels] Unexpected error: %v", err)
			} else {
				fmt.Println("[TestRunTasksWithChannels] Test passed: No errors")
			}
		case <-time.After(1 * time.Second):
			fmt.Println("[TestRunTasksWithChannels] Test passed: Timeout occurred, no errors detected")
		}
	})

	t.Run("Task panics and propagates error", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithChannels] Starting test: Task panics and propagates error")

		tasks := 5
		errChan := make(chan error, 1)

		go RunTasksWithChannels(tasks, mockTask, errChan, 2) // Panic on task 2

		select {
		case err := <-errChan:
			if err == nil {
				t.Error("[TestRunTasksWithChannels] Expected error, got nil")
			} else {
				fmt.Printf("[TestRunTasksWithChannels] Test passed: Received expected error: %v\n", err)
			}
		case <-time.After(1 * time.Second):
			t.Error("[TestRunTasksWithChannels] Timeout waiting for error")
		}
	})
}
