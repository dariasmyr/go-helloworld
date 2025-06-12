package worker_pool

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

const (
	defaultMaxWorkers = 10
	defaultDur        = 250 * time.Millisecond
)

type Task struct{ ID int }
type Res struct {
	taskId int
	status bool
}

type SimpleWorkerPool struct {
	tasks      chan Task
	res        chan Res
	dur        time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	maxWorkers int
}

func simpleWorker(ctx context.Context, tasks <-chan Task, results chan<- Res) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tasks:
			if !ok {
				fmt.Printf("Chan tasks closed\n")
				return
			}
			fmt.Printf("Processing task %d\n", task.ID)
			results <- Res{
				taskId: task.ID,
				status: true,
			}
		}
	}
}

func NewWorkerPool(maxWorkers int, dur time.Duration, maxQueueSize int, ctx context.Context, cancel context.CancelFunc) *SimpleWorkerPool {
	if dur == 0 {
		dur = defaultDur
	}

	if maxWorkers == 0 {
		maxWorkers = defaultMaxWorkers
	}

	return &SimpleWorkerPool{
		tasks:      make(chan Task, maxQueueSize),
		res:        make(chan Res),
		dur:        dur,
		maxWorkers: maxWorkers,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (wp *SimpleWorkerPool) Run(wg *sync.WaitGroup) {
	for i := 0; i < wp.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			simpleWorker(wp.ctx, wp.tasks, wp.res)
		}()
	}
}

func TestWorkerPool(t *testing.T) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wp := NewWorkerPool(0, 0, 10, ctx, cancel)

	wp.Run(&wg)

	go wp.readRes()
	go func() {
		wg.Wait()
		close(wp.res)
	}()

	for i := 0; i <= 50; i++ {
		AddTask(ctx, i, wp.tasks, wp.dur)
	}
	close(wp.tasks)
}

func (wp *SimpleWorkerPool) Stop() {
	wp.cancel()
}

func (wp *SimpleWorkerPool) readRes() {
	for {
		select {
		case <-wp.ctx.Done():
			fmt.Printf("Context closed\n")
			return
		case res, ok := <-wp.res:
			if !ok {
				fmt.Printf("Res chan closed\n")
				return
			}
			fmt.Printf("Res %d; status %v\n", res.taskId, res.status)
		}
	}
}

func AddTask(ctx context.Context, n int, tasks chan<- Task, dur time.Duration) {
	select {
	case <-ctx.Done():
		fmt.Printf("Context closed\n")
		return
	case tasks <- Task{ID: n}:
		fmt.Printf("Task %d sent to tasks chan\n", n)
	case <-time.After(dur):
		fmt.Printf("Timer ticked, skip task sending\n")
		return
	}
}
