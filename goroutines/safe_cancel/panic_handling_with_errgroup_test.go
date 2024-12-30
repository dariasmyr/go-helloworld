package goroutines_safety

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/sync/errgroup"
)

// RunTasksWithErrgroup runs multiple tasks in parallel using errgroup.WithContext() for cancellation.
// If one of the tasks causes a panic, it is handled by errgrupd (as it shared common context and shows error trace).
// tasks - the number of tasks, taskFunc - a function to perform, failOn - the number of a task that causes panic.
func RunTasksWithErrgroup(tasks int, taskFunc func(int, int) func(), failOn int) error {
	fmt.Println("[RunTasksWithErrgroup] Starting tasks...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < tasks; i++ {
		g.Go(func() error {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("[RunTasksWithErrgroup] Recovered from panic:", r)
					cancel()
				}
			}()
			fmt.Printf("[RunTasksWithErrgroup] Task %d started\n", i)
			taskFunc(i, failOn)()
			fmt.Printf("[RunTasksWithErrgroup] Task %d completed\n", i)

			select {
			case <-ctx.Done():
				return fmt.Errorf("task %d canceled: %v", i, ctx.Err())
			default:
				return nil
			}
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	fmt.Println("[RunTasksWithErrgroup] All tasks completed.")
	return nil
}

func TestRunTasksWithErrgroup(t *testing.T) {
	t.Run("All tasks complete without panic", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithErrgroup] Starting test: All tasks complete without panic")

		tasks := 5

		err := RunTasksWithErrgroup(tasks, MockTask, -1)

		if err != nil {
			t.Errorf("[TestRunTasksWithErrgroup] Unexpected error: %v", err)
		} else {
			fmt.Println("[TestRunTasksWithErrgroup] Test passed: No errors")
		}
	})

	t.Run("Task panics and propagates error", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithErrgroup] Starting test: Task panics and propagates error")

		tasks := 5

		err := RunTasksWithErrgroup(tasks, MockTask, 2)

		if err == nil {
			t.Error("[TestRunTasksWithErrgroup] Expected error, got nil")
		} else {
			fmt.Printf("[TestRunTasksWithErrgroup] Test passed: Received expected error: %v\n", err)
		}
	})
}
