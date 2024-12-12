package worker_pool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func slowAddition(ctx context.Context, args interface{}) (interface{}, error) {
	nums, ok := args.([]int)
	if !ok || len(nums) != 2 {
		return nil, fmt.Errorf("invalid args")
	}
	time.Sleep(100 * time.Millisecond) // Mock long task
	return nums[0] + nums[1], nil
}

func TestWorkerPool_Success(t *testing.T) {
	jobs := []Job{
		{
			Description: JobDescriptor{
				ID:      "job1",
				JobType: "addition",
			},
			ExecFn: slowAddition,
			Args:   []int{1, 2},
		},
		{
			Description: JobDescriptor{
				ID:      "job2",
				JobType: "addition",
			},
			ExecFn: slowAddition,
			Args:   []int{3, 4},
		},
	}

	pool := New(2)

	pool.GenerateFrom(jobs)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go pool.Run(ctx)

	results := make([]Result, len(jobs))
	i := 0
	for result := range pool.results {
		results[i] = result
		i++
	}

	for _, result := range results {
		if result.Err != nil {
			t.Errorf("expected no error, got %v", result.Err)
		}
		if result.Value == nil {
			t.Errorf("expected a value, got nil")
		}
	}
}

func TestWorkerPool_Cancel(t *testing.T) {
	job := Job{
		Description: JobDescriptor{ID: "job1", JobType: "addition"},
		ExecFn:      slowAddition,
		Args:        []int{1, 2},
	}

	pool := New(1)

	pool.GenerateFrom([]Job{job})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	go pool.Run(ctx)

	select {
	case result := <-pool.results:
		if result.Err == nil {
			fmt.Printf("expected an error due to context cancelation, but got %v", result.Value)
		}
	case <-time.After(time.Millisecond * 150):
		t.Errorf("test timed out, result not received")
	}
}
