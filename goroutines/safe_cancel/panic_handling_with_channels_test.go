package goroutines_safety

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// RunTasksWithChannels runs multiple tasks in parallel.
// If one of the tasks causes a panic, it is processed with recover() and passed to errChan.
// tasks - the number of tasks, taskFunc - a function to perform, errChan - a channel for errors, failOn - the number of a task that causes panic.
func RunTasksWithChannels(tasks int, taskFunc func(int, int) func(), errChan chan error, failOn int) error {
	fmt.Println("[RunTasksWithChannels] Starting tasks...")

	var wg sync.WaitGroup

	var err error

	go func(errChan chan error, e *error) {
		for {
			select {
			case err := <-errChan:
				if err != nil {
					*e = err
				}
			}
		}
	}(errChan, &err)

	wg.Add(tasks)
	for i := 0; i < tasks; i++ {
		go func(i int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicError := fmt.Errorf("task %d panicked: %v", i, r)
					fmt.Println("[RunTasksWithChannels] Recovered from panic:", panicError)
					errChan <- panicError
				}
			}()

			if err == nil {
				taskFunc(i, failOn)()
			}
		}(i)
	}

	wg.Wait()

	fmt.Println("[RunTasksWithChannels] All tasks completed. Closing error channel.")
	close(errChan)
	return err
}

func MockTask(ctx context.Context, id int, failOn int) func() {
	return func() {
		select {
		case <-ctx.Done():
			fmt.Println("[MockTask] Context is done, do not continue executing task ", id)
			return
		default:
			fmt.Printf("[mockTask] Task %d start executing\n", id)
			if id == failOn {
				fmt.Printf("[mockTask] Task %d will panic\n", id)
				panic("Task failed")
			} else {
				time.Sleep(100 * time.Millisecond) // Mock executing
				fmt.Printf("[mockTask] Task %d finished executing\n", id)
			}
		}
	}
}

func MockTaskForChannels(id int, failOn int) func() {
	return func() {
		fmt.Printf("[mockTaskForChannels] Task %d start executing\n", id)
		if id == failOn {
			fmt.Printf("[mockTaskForChannels] Task %d will panic\n", id)
			panic("Task failed")
		} else {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("[mockTaskForChannels] Task %d finished executing\n", id)
		}
	}
}

func TestRunTasksWithChannels(t *testing.T) {
	t.Run("All tasks complete without panic", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithChannels] Starting test: All tasks complete without panic")

		tasks := 5
		errChan := make(chan error, 1)

		err := RunTasksWithChannels(tasks, MockTaskForChannels, errChan, -1) // No errors (failOn=-1)

		if err != nil {
			t.Errorf("[TestRunTasksWithChannels] Unexpected error: %v", err)
		} else {
			fmt.Println("[TestRunTasksWithChannels] Test passed: No errors")
		}

	})

	t.Run("Task panics and propagates error", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithChannels] Starting test: Task panics and propagates error")

		tasks := 5
		errChan := make(chan error, 1)

		err := RunTasksWithChannels(tasks, MockTaskForChannels, errChan, 2) // Panic on task 2

		if err == nil {
			t.Error("[TestRunTasksWithChannels] Expected error, got nil")
		} else {
			fmt.Printf("[TestRunTasksWithChannels] Test passed: Received expected error: %v\n", err)
		}
	})
}
