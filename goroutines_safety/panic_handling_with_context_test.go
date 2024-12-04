package goroutines_safety

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// RunTasksWithContext runs multiple tasks in parallel.
// If one of the tasks causes a panic, it cancels the context to stop all other tasks.
// tasks - the number of tasks, taskFunc - a function to perform, failOn - the number of a task that causes panic and cancel context.
func RunTasksWithContext(ctx context.Context, tasks int, taskFunc func(int, int, context.Context) func(), failOn int) {
	fmt.Println("[RunTasksWithContext] Starting tasks...")

	var wg sync.WaitGroup
	wg.Add(tasks)

	ctx, cancel := context.WithCancel(ctx) // Create child context
	defer cancel()                         // Close context after func completing

	for i := 0; i < tasks; i++ {
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("task %d panicked: %v", id, r)
					fmt.Println("[RunTasksWithContext] Recovered from panic:", err)
					cancel() // Close all other tasks
				}
			}()

			fmt.Printf("[RunTasksWithContext] Task %d started\n", id)
			taskFunc(id, failOn, ctx)()
			fmt.Printf("[RunTasksWithContext] Task %d completed\n", id)
		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish
	fmt.Println("[RunTasksWithContext] All tasks completed")
}

func mockTaskWithContext(id int, failOn int, ctx context.Context) func() {
	return func() {
		select {
		case <-ctx.Done(): // Проверяем сигнал отмены
			fmt.Printf("[mockTask] Task %d cancelled due to context\n", id)
			return
		default:
			fmt.Printf("[mockTask] Task %d executing\n", id)
			if id == failOn {
				fmt.Printf("[mockTask] Task %d will panic\n", id)
				panic(fmt.Sprintf("Task %d failed", id))
			}
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("[mockTask] Task %d execution finished\n", id)
		}
	}
}

func TestRunTasksWithContext(t *testing.T) {
	t.Run("All tasks complete without panic", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithContext] Starting test: All tasks complete without panic")

		tasks := 5

		ctx := context.Background()                                 // Root context
		go RunTasksWithContext(ctx, tasks, mockTaskWithContext, -1) // No errors (failOn=-1)

		select {
		case <-time.After(1 * time.Second):
			fmt.Println("[TestRunTasksWithContext] Test passed: Timeout occurred, no errors detected")
		}
	})

	t.Run("Task panics and propagates error", func(t *testing.T) {
		fmt.Println("[TestRunTasksWithContext] Starting test: Task panics and propagates error")

		tasks := 5

		ctx := context.Background()                                // Root context
		go RunTasksWithContext(ctx, tasks, mockTaskWithContext, 2) // Panic on task 2

		select {
		case <-time.After(1 * time.Second):
			t.Error("[TestRunTasksWithContext] Timeout waiting for error")
		}
	})
}
