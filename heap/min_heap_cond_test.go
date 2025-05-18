package heap

import (
	"container/heap"
	"fmt"
	"sync"
	"testing"
	"time"
)

type Job struct {
	ID        int
	Priority  int // lower number is higher priority
	Timestamp time.Time
}

type JobPriorityQueue []*Job

func (pq JobPriorityQueue) Len() int {
	return len(pq)
}

func (pq JobPriorityQueue) Less(i, j int) bool {
	if pq[i].Priority == pq[j].Priority {
		return pq[i].Timestamp.Before(pq[j].Timestamp)
	}
	return pq[i].Priority < pq[j].Priority
}

func (pq JobPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *JobPriorityQueue) Push(x interface{}) {
	item := x.(*Job)
	*pq = append(*pq, item)
}

func (pq *JobPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	lastElem := old[n-1]
	*pq = old[:n-1]
	return lastElem
}

type SafeJobQueue struct {
	mu       sync.Mutex
	pq       JobPriorityQueue
	cond     *sync.Cond
	isClosed bool
}

func NewSafeJobQueue() *SafeJobQueue {
	q := &SafeJobQueue{
		pq: make(JobPriorityQueue, 0),
	}
	q.cond = sync.NewCond(&q.mu)
	heap.Init(&q.pq)
	return q
}

func (q *SafeJobQueue) Push(job *Job) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.isClosed {
		return
	}

	heap.Push(&q.pq, job)
	q.cond.Signal()
}

func (q *SafeJobQueue) Pop() *Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	for q.pq.Len() == 0 && !q.isClosed {
		q.cond.Wait()
	}

	if q.pq.Len() == 0 && q.isClosed {
		return nil
	}

	fmt.Println("Unblocking job queue")
	return heap.Pop(&q.pq).(*Job)
}

func (q *SafeJobQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.isClosed = true
	q.cond.Broadcast()
}

func worker(id int, q *SafeJobQueue, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		job := q.Pop()
		if job == nil {
			fmt.Printf("worker [%d] exiting\n", id)
			return
		}
		fmt.Printf("worker [%d] Handling job [%d] (priority %d)\n", id, job.ID, job.Priority)
		time.Sleep(1 * time.Second)

	}
}

func TestHeapCondQueue(t *testing.T) {
	q := NewSafeJobQueue()
	var wg sync.WaitGroup

	for i := 0; i <= 5; i++ {
		wg.Add(1)
		go worker(i, q, &wg)
	}

	for i := 0; i < 10; i++ {
		job := &Job{
			ID:        i,
			Priority:  10 - i%5,
			Timestamp: time.Now(),
		}
		q.Push(job)
		time.Sleep(500 * time.Millisecond)
	}

	q.Close()

	wg.Wait()
}
