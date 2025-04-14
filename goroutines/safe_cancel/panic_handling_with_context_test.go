package goroutines_safety

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// RunTasksWithContext runs multiple tasks in parallel using context for cancellation.
// If one of the tasks causes a panic, it is processed with recover() and added to the context error.
// tasks - the number of tasks, taskFunc - a function to perform, failOn - the number of a task that causes panic.
func RunTasksWithContext(ctx context.Context, tasks int, taskFunc func(context.Context, int, int) func(), failOn int) error {
	fmt.Println("[RunTasksWithContext] Starting tasks...")

	var wg sync.WaitGroup
	var mu sync.Mutex
	var err error

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(tasks)

	for i := 0; i < tasks; i++ {
		go func(i int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					err = fmt.Errorf("task %d panicked: %v", i, r)
					mu.Unlock()

					cancel()
					fmt.Println("[RunTasksWithContext] Recovered from panic:", r)
				}
			}()

			select {
			case <-ctx.Done():
				fmt.Println("[RunTasksWithContext] Context is done, do not start task: ", i)
				return
			default:
				taskFunc(ctx, i, failOn)()
			}
		}(i)
	}

	wg.Wait()

	if err != nil {
		return err
	}

	fmt.Println("[RunTasksWithContext] All tasks completed.")
	return nil
}

func TestRunTasksWithContext(t *testing.T) {
	t.Run("All tasks complete without panic", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithContext] Starting test: All tasks complete without panic")

		tasks := 5
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := RunTasksWithContext(ctx, tasks, MockTask, -1)

		if err != nil {
			t.Errorf("[TestRunTasksWithContext] Unexpected error: %v", err)
		} else {
			fmt.Println("[TestRunTasksWithContext] Test passed: No errors")
		}
	})

	t.Run("Task panics and propagates error", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithContext] Starting test: Task panics and propagates error")

		tasks := 10
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := RunTasksWithContext(ctx, tasks, MockTask, 2)

		if err == nil {
			t.Error("[TestRunTasksWithContext] Expected error, got nil")
		} else {
			fmt.Printf("[TestRunTasksWithContext] Test passed: Received expected error: %v\n", err)
		}
	})
}
