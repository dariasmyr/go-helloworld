package semaphore

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

// TryAcquire пытается захватить слот и возвращает true при успехе.
// Реализация:
//  1. fast check load -> отказ если уже >= limit
//  2. короткий спин-цикл (spinCount итераций) с CAS + небольшая
//     пауза runtime.Gosched() для снижения горения CPU
//  3. если spin не дал результата -> exponential backoff (time.Sleep)
func (s *Sem) OptimizedTryAcquire() bool {
	// Быстрая проверка: если уже заполнено — сразу отказ
	if s.cur.Load() >= s.limit {
		return false
	}

	// Короткий спин с CAS — уменьшает количество Add/rollback и конкуренцию
	const spinCount = 302
	for i := 0; i < spinCount; i++ {
		cur := s.cur.Load()
		if cur >= s.limit {
			return false
		}
		if s.cur.CompareAndSwap(cur, cur+1) {
			return true
		}
		// периодически уступаем планировщику — снижает горячую петлю
		// (не делаем Sleep, чтобы не терять высокопроизводительный fast-path)
		if (i & 7) == 0 {
			runtime.Gosched()
		}
	}

	// Экспоненциальный backoff
	backoff := 1 * time.Microsecond
	const maxBackoff = 50 * time.Millisecond
	for {
		// до сна проверяем — возможно место освободилось
		cur := s.cur.Load()
		if cur >= s.limit {
			return false
		}
		if s.cur.CompareAndSwap(cur, cur+1) {
			return true
		}
		time.Sleep(backoff)
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
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

func (s *Sem) Release() {
	cur := s.cur.Add(-1)
	if cur < 0 {
		panic("semaphore release without acquire")
	}
}

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

func BenchmarkSimpleAtomicSemParams(b *testing.B) {
	limits := powersOf(8, 4096, 8)
	goroutinesList := powersOf(64, 32768, 8)

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

func BenchmarkOptimizedAtomicSemParams(b *testing.B) {
	limits := powersOf(8, 4096, 8)
	goroutinesList := powersOf(64, 32768, 8)

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
								for !sem.OptimizedTryAcquire() {
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
	limits := powersOf(8, 4096, 8)           // 8,16,...,1024
	goroutinesList := powersOf(64, 32768, 8) // 64,128,...,8192

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
