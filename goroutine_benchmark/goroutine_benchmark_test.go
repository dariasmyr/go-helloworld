package goroutine_benchmark

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"
)

func cpuTask(iterations int) {
	sum := 0
	for i := 0; i < iterations; i++ {
		sum += rand.Intn(1000)
	}
	_ = sum
}

func ioTask(latency time.Duration) {
	time.Sleep(latency)
}

type resourceUsage struct {
	userTime   time.Duration
	systemTime time.Duration
	maxRSSKB   int64
	minFaults  int64
	majFaults  int64
	inBlocks   int64
	outBlocks  int64
}

func getResourceUsage() resourceUsage {
	var ru syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
	return resourceUsage{
		userTime:   time.Duration(ru.Utime.Sec)*time.Second + time.Duration(ru.Utime.Usec)*time.Microsecond,
		systemTime: time.Duration(ru.Stime.Sec)*time.Second + time.Duration(ru.Stime.Usec)*time.Microsecond,
		maxRSSKB:   ru.Maxrss,  // KB
		minFaults:  ru.Minflt,  // soft page faults
		majFaults:  ru.Majflt,  // hard page faults
		inBlocks:   ru.Inblock, // disk reads
		outBlocks:  ru.Oublock, // disk writes
	}
}

func measureCPUPercent(start resourceUsage, elapsed time.Duration) float64 {
	end := getResourceUsage()
	deltaUser := end.userTime - start.userTime
	deltaSys := end.systemTime - start.systemTime
	totalCPU := deltaUser + deltaSys
	return (float64(totalCPU) / float64(elapsed)) * 100
}

func runBenchmark(numWorkers int, workType string, iterations int, latency time.Duration) (time.Duration, float64, int, resourceUsage) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	startUsage := getResourceUsage()
	start := time.Now()

	var wg sync.WaitGroup
	tasks := make(chan struct{}, iterations)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range tasks {
				if workType == "cpu" {
					cpuTask(50_000)
				} else if workType == "io" {
					ioTask(latency)
				}
			}
		}()
	}

	for i := 0; i < iterations; i++ {
		tasks <- struct{}{}
	}
	close(tasks)
	wg.Wait()

	elapsed := time.Since(start)
	cpuPercent := measureCPUPercent(startUsage, elapsed)
	endUsage := getResourceUsage()
	return elapsed, cpuPercent, runtime.NumGoroutine(), endUsage
}

func Test(t *testing.T) {

	for _, workType := range []string{"cpu", "io"} {
		for workers := 1; workers <= 64; workers *= 2 {
			elapsed, cpuPercent, goroutines, usage := runBenchmark(
				workers,
				workType,
				1000,                // task number
				50*time.Millisecond, // sleep time for i/o imitation
			)

			fmt.Printf("[%s] Workers=%d → Time=%v, CPU=%.2f%%, Goroutines=%d, MaxRSS=%d KB, SoftPF=%d, HardPF=%d\n",
				workType, workers, elapsed, cpuPercent, goroutines,
				usage.maxRSSKB, usage.minFaults, usage.majFaults)
		}
	}
}
