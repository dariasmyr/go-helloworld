package worker_pool

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func addNumbers(ctx context.Context, args interface{}) (interface{}, error) {
	nums, ok := args.([]int)
	if !ok || len(nums) != 2 {
		return nil, fmt.Errorf("invalid agrumens")
	}

	time.Sleep(100 * time.Millisecond) // Mock logic execution

	return nums[0] + nums[1], nil
}

func TestJobExecution_Success(t *testing.T) {
	job := Job{
		Description: JobDescriptor{
			ID:      "job1",
			JobType: "addition",
		},
		ExecFn: addNumbers,
		Args:   []int{1, 2},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result := job.execute(ctx)

	if result.Err != nil {
		t.Errorf("expected no error, got %v", result.Err)
	}

	expectedValue := 3
	if result.Value != expectedValue {
		t.Errorf("expected value %v, got %v", expectedValue, result.Value)
	}
}

func TestJobExecution_Failure(t *testing.T) {
	job := Job{
		Description: JobDescriptor{
			ID:      "job2",
			JobType: "addition",
		},
		ExecFn: addNumbers,
		Args:   []int{1},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result := job.execute(ctx)

	if result.Err == nil {
		t.Errorf("expected an error, but got none")
	}
	if result.Value != nil {
		t.Errorf("expected no result value, got %v", result.Value)
	}
}
