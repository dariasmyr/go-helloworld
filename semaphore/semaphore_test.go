package semaphore

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

type Sem struct {
	limit int32
	cur   *atomic.Int32
}

func NewSem(init int) *Sem {
	return &Sem{
		limit: int32(init),
		cur:   &atomic.Int32{},
	}
}

func (s *Sem) TryAcquire() bool {
	for {
		cur := s.cur.Load()
		if cur >= s.limit {
			return false
		}
		if s.cur.CompareAndSwap(cur, cur+1) {
			return true
		}
	}
}

func (s *Sem) Release() { s.cur.Add(-1) }

func doWork() int {
	arr := make([]int, 100)
	sum := 0
	for i := range arr {
		arr[i] = rand.Intn(100)
		sum += arr[i] * (i + 1)
	}
	return sum
}

func powersOf(min, max, step int) []int {
	if step <= 1 {
		panic("step must be > 1")
	}
	res := []int{}
	for val := min; val <= max; val *= step {
		res = append(res, val)
	}
	return res
}

var (
	limit      = 50
	goroutines = 200
)

func BenchmarkAtomicSem(b *testing.B) {
	for b.Loop() {
		sem := NewSem(limit)
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for g := 0; g < goroutines; g++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 1000; j++ {
					sem.TryAcquire()
					doWork()
					sem.Release()
				}
			}()
		}

		wg.Wait()
	}
}

func BenchmarkChannelSem(b *testing.B) {
	for b.Loop() {
		sem := make(chan struct{}, limit)
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for g := 0; g < goroutines; g++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 1000; j++ {
					sem <- struct{}{}
					doWork()
					<-sem
				}
			}()
		}
		wg.Wait()
	}
}

func BenchmarkAtomicSemParams(b *testing.B) {
	limits := powersOf(4, 1024, 16)             // 8,16,...,1024
	goroutinesList := powersOf(64, 262_144, 16) // 64,128,...,8192

	for _, limit := range limits {
		for _, goroutines := range goroutinesList {
			b.Run(fmt.Sprintf("Atomic/Limit=%d/Goroutines=%d", limit, goroutines), func(b *testing.B) {
				for b.Loop() {
					sem := NewSem(limit)
					var wg sync.WaitGroup
					wg.Add(goroutines)
					for g := 0; g < goroutines; g++ {
						go func() {
							defer wg.Done()
							for j := 0; j < 1000; j++ {
								for !sem.TryAcquire() {
								}
								doWork()
								sem.Release()
							}
						}()
					}
					wg.Wait()
				}
			})
		}
	}
}

func BenchmarkChannelSemParams(b *testing.B) {
	limits := powersOf(4, 1024, 16)
	goroutinesList := powersOf(64, 262_144, 16)

	for _, limit := range limits {
		for _, goroutines := range goroutinesList {
			b.Run(fmt.Sprintf("Channel/Limit=%d/Goroutines=%d", limit, goroutines), func(b *testing.B) {
				for b.Loop() {
					sem := make(chan struct{}, limit)
					var wg sync.WaitGroup
					wg.Add(goroutines)
					for g := 0; g < goroutines; g++ {
						go func() {
							defer wg.Done()
							for j := 0; j < 1000; j++ {
								sem <- struct{}{}
								doWork()
								<-sem
							}
						}()
					}
					wg.Wait()
				}
			})
		}
	}
}
