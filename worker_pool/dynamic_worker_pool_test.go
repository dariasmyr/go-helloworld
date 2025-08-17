package worker_pool

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func runLoad(workers int, requestsPerWorker int, cpuBound bool) (throughput float64, p50, p95, p99 float64) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allLatencies []float64

	start := time.Now()

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			latencies := make([]float64, 0, requestsPerWorker)

			for j := 0; j < requestsPerWorker; j++ {
				reqStart := time.Now()

				if cpuBound {
					// CPU-bound операция
					_ = fibonacci(35) // пример трудоёмкой операции
				} else {
					// I/O-bound операция (HTTP запрос)
					resp, err := client.Get("http://localhost:8080/echo?echo=" +
						fmt.Sprintf("hello_worker_%d_request_%d", workerID, j))
					if err != nil {
						fmt.Println("Request error:", err)
						continue
					}

					_, err = io.ReadAll(resp.Body)
					resp.Body.Close()
					if err != nil {
						fmt.Println("Read body error:", err)
						continue
					}
				}

				elapsed := time.Since(reqStart).Seconds() * 1000 // ms
				latencies = append(latencies, elapsed)
			}

			// объединяем с общим срезом под мьютексом
			mu.Lock()
			allLatencies = append(allLatencies, latencies...)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	totalTime := time.Since(start).Seconds()
	totalRequests := workers * requestsPerWorker
	throughput = float64(totalRequests) / totalTime

	// сортируем для перцентилей
	sort.Float64s(allLatencies)
	n := len(allLatencies)
	getPercentile := func(p float64) float64 {
		k := int(math.Ceil(float64(n)*p)) - 1
		if k < 0 {
			k = 0
		}
		return allLatencies[k]
	}

	p50 = getPercentile(0.50)
	p95 = getPercentile(0.95)
	p99 = getPercentile(0.99)

	return
}

func TestSaturationPoint(t *testing.T) {
	fmt.Printf("CPU cores: %d | GOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))

	requestsPerWorker := 20
	for _, workers := range []int{1, 2, 5, 10, 20, 50, 100, 200} {
		throughput, p50, p95, p99 := runLoad(workers, requestsPerWorker, false)
		fmt.Printf("Workers=%d | Throughput=%.2f req/s | p50=%.1fms | p95=%.1fms | p99=%.1fms\n",
			workers, throughput, p50, p95, p99)
	}
}

// CPU Bound fibonacci task
///usr/local/go/bin/go tool test2json -t /Users/darya.smyr/Library/Caches/JetBrains/GoLand2023.2/tmp/GoLand/___TestSaturationPoint_in_go_helloworld_worker_pool.test -test.v -test.paniconexit0 -test.run ^\QTestSaturationPoint\E$
//=== RUN   TestSaturationPoint
//CPU cores: 11 | GOMAXPROCS: 11
//Workers=1 | Throughput=37.32 req/s | p50=26.4ms | p95=28.1ms | p99=37.6ms
//Workers=2 | Throughput=74.78 req/s | p50=26.6ms | p95=27.1ms | p99=27.4ms
//Workers=5 | Throughput=180.35 req/s | p50=27.7ms | p95=28.0ms | p99=28.1ms
//Workers=10 | Throughput=275.55 req/s | p50=35.3ms | p95=41.8ms | p99=44.6ms
//Workers=20 | Throughput=284.33 req/s | p50=63.0ms | p95=98.0ms | p99=105.7ms
//Workers=50 | Throughput=290.83 req/s | p50=147.9ms | p95=283.3ms | p99=340.4ms
//Workers=100 | Throughput=291.71 req/s | p50=263.3ms | p95=572.6ms | p99=671.1ms
//Workers=200 | Throughput=287.98 req/s | p50=593.0ms | p95=1195.5ms | p99=1336.0ms
//--- PASS: TestSaturationPoint (27.94s)
//PASS

// IO Bound task without time sleep
//=== RUN   TestSaturationPoint
//CPU cores: 11 | GOMAXPROCS: 11
//Workers=1 | Throughput=1970.17 req/s | p50=0.3ms | p95=0.3ms | p99=5.1ms
//Workers=2 | Throughput=7579.23 req/s | p50=0.2ms | p95=0.3ms | p99=1.3ms
//Workers=5 | Throughput=9275.36 req/s | p50=0.5ms | p95=0.8ms | p99=1.2ms
//Workers=10 | Throughput=12374.48 req/s | p50=0.7ms | p95=0.9ms | p99=2.2ms
//Workers=20 | Throughput=28868.97 req/s | p50=0.6ms | p95=1.1ms | p99=2.0ms
//Workers=50 | Throughput=32091.74 req/s | p50=1.5ms | p95=2.7ms | p99=4.6ms
//Workers=100 | Throughput=32102.71 req/s | p50=2.9ms | p95=5.2ms | p99=9.4ms
//Workers=200 | Throughput=27746.03 req/s | p50=6.2ms | p95=12.0ms | p99=17.6ms
//--- PASS: TestSaturationPoint (0.29s)

// IO Bound task + 100 ms
//CPU cores: 11 | GOMAXPROCS: 11
//Workers=1 | Throughput=9.77 req/s | p50=102.4ms | p95=102.8ms | p99=105.3ms
//Workers=2 | Throughput=19.53 req/s | p50=102.3ms | p95=102.8ms | p99=104.4ms
//Workers=5 | Throughput=48.80 req/s | p50=102.4ms | p95=103.0ms | p99=104.4ms
//Workers=10 | Throughput=97.32 req/s | p50=102.7ms | p95=103.9ms | p99=106.3ms
//Workers=20 | Throughput=193.02 req/s | p50=103.4ms | p95=105.3ms | p99=108.6ms
//Workers=50 | Throughput=484.67 req/s | p50=102.2ms | p95=106.4ms | p99=117.1ms
//Workers=100 | Throughput=966.73 req/s | p50=102.5ms | p95=105.2ms | p99=125.0ms
//Workers=200 | Throughput=1841.05 req/s | p50=102.3ms | p95=111.6ms | p99=200.9ms
//--- PASS: TestSaturationPoint (16.58s)

// IO Bound task + 200 ms
//=== RUN   TestSaturationPoint
//CPU cores: 11 | GOMAXPROCS: 11
//Workers=1 | Throughput=4.94 req/s | p50=202.3ms | p95=202.8ms | p99=205.8ms
//Workers=2 | Throughput=9.89 req/s | p50=202.2ms | p95=202.7ms | p99=202.9ms
//Workers=5 | Throughput=24.70 req/s | p50=202.4ms | p95=203.1ms | p99=204.8ms
//Workers=10 | Throughput=49.33 req/s | p50=202.5ms | p95=204.6ms | p99=206.0ms
//Workers=20 | Throughput=98.24 req/s | p50=203.3ms | p95=205.1ms | p99=209.0ms
//Workers=50 | Throughput=245.85 req/s | p50=202.8ms | p95=205.8ms | p99=214.8ms
//Workers=100 | Throughput=491.39 req/s | p50=202.6ms | p95=208.2ms | p99=224.2ms
//Workers=200 | Throughput=978.44 req/s | p50=202.6ms | p95=211.0ms | p99=243.9ms
//--- PASS: TestSaturationPoint (32.49s)
//PASS

type TestTask func()

func fixedWorkerPool(tasks []TestTask, workers int) {
	tasksCh := make(chan TestTask)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tasksCh {
				t()
			}
		}()
	}

	for _, t := range tasks {
		tasksCh <- t
		time.Sleep(2 * time.Millisecond)
	}
	close(tasksCh)
	wg.Wait()
}

type DynamicWorkerPool struct {
	jobs        chan TestTask
	minWorkers  int
	maxWorkers  int
	idleTimeout time.Duration

	mu       sync.Mutex
	workers  int
	wg       sync.WaitGroup
	shutdown chan struct{}
}

func NewDynamicWorkerPool(minWorkers, maxWorkers int, idleTimeout time.Duration) *DynamicWorkerPool {
	pool := &DynamicWorkerPool{
		jobs:        make(chan TestTask),
		minWorkers:  minWorkers,
		maxWorkers:  maxWorkers,
		idleTimeout: idleTimeout,
		shutdown:    make(chan struct{}),
	}
	for i := 0; i < minWorkers; i++ {
		pool.startWorker()
	}
	return pool
}

func (p *DynamicWorkerPool) startWorker() {
	p.mu.Lock()
	if p.workers >= p.maxWorkers {
		p.mu.Unlock()
		return
	}
	p.workers++
	p.mu.Unlock()

	p.wg.Add(1)
	go func() {
		defer func() {
			p.mu.Lock()
			p.workers--
			p.mu.Unlock()
			p.wg.Done()
		}()

		timer := time.NewTimer(p.idleTimeout)
		defer timer.Stop()

		for {
			select {
			case task := <-p.jobs:
				if task != nil {
					task()
				}
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(p.idleTimeout)
			case <-timer.C:
				return
			case <-p.shutdown:
				return
			}
		}
	}()
}

func (p *DynamicWorkerPool) Submit(task TestTask) {
	select {
	case p.jobs <- task:
	default:
		p.startWorker()
		p.jobs <- task
	}
}

func (p *DynamicWorkerPool) Stop() {
	close(p.shutdown)
	p.wg.Wait()
}

func TestExecutors(t *testing.T) {
	const numTasks = 200
	fmt.Println("Num tasks: fibonacci(20) ", numTasks)

	task := func() {
		fibonacci(20)
	}

	tasks := make([]TestTask, numTasks)
	for i := 0; i < numTasks; i++ {
		tasks[i] = task
	}

	measure := func(name string, exec func()) {
		start := time.Now()
		exec()
		elapsed := time.Since(start)
		fmt.Printf("%s: elapsed=%v, goroutines=%d\n",
			name, elapsed, runtime.NumGoroutine())
	}

	t.Run("Fixed pool", func(t *testing.T) {
		workers := runtime.NumCPU() * 2
		measure("Fixed", func() {
			fixedWorkerPool(tasks, workers)
		})
	})

	t.Run("Dynamic pool with gradual submission", func(t *testing.T) {
		pool := NewDynamicWorkerPool(2, runtime.NumCPU()*2, 50*time.Millisecond)
		measure("Dynamic", func() {
			for _, t := range tasks {
				pool.Submit(t)
				time.Sleep(2 * time.Millisecond) // постепенная подача
			}
			pool.Stop()
		})
	})
}

//=== RUN   TestExecutors
//Num tasks: fibonacci(20)  200
//--- PASS: TestExecutors (0.91s)
//=== RUN   TestExecutors/Fixed_pool
//Fixed: elapsed=454.153708ms, goroutines=3
//--- PASS: TestExecutors/Fixed_pool (0.45s)
//=== RUN   TestExecutors/Dynamic_pool_with_gradual_submission
//Dynamic: elapsed=456.418542ms, goroutines=3
//--- PASS: TestExecutors/Dynamic_pool_with_gradual_submission (0.46s)
//PASS

// ----------------------
// Динамический семафор
// ----------------------
type Sem struct {
	limit int32
	cur   int32
}

func NewSem(init int) *Sem { return &Sem{limit: int32(init)} }
func (s *Sem) TryAcquire() bool {
	for {
		cur := atomic.LoadInt32(&s.cur)
		lim := atomic.LoadInt32(&s.limit)
		if cur >= lim {
			return false
		}
		if atomic.CompareAndSwapInt32(&s.cur, cur, cur+1) {
			return true
		}
	}
}
func (s *Sem) Release()       { atomic.AddInt32(&s.cur, -1) }
func (s *Sem) SetLimit(n int) { atomic.StoreInt32(&s.limit, int32(n)) }
func (s *Sem) Limit() int     { return int(atomic.LoadInt32(&s.limit)) }
func (s *Sem) Cur() int       { return int(atomic.LoadInt32(&s.cur)) }

// ----------------------
// Простое окно и p95
// ----------------------
type Window struct {
	mu   sync.Mutex
	buf  []float64
	pos  int
	full bool
}

func NewWindow(n int) *Window { return &Window{buf: make([]float64, n)} }
func (w *Window) Add(v float64) {
	w.mu.Lock()
	w.buf[w.pos] = v
	w.pos++
	if w.pos >= len(w.buf) {
		w.pos = 0
		w.full = true
	}
	w.mu.Unlock()
}
func (w *Window) Snapshot() []float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := w.pos
	if w.full {
		n = len(w.buf)
	}
	out := make([]float64, n)
	copy(out, w.buf[:n])
	return out
}
func percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sort.Float64s(xs)
	idx := int(float64(len(xs))*p+0.5) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(xs) {
		idx = len(xs) - 1
	}
	return xs[idx]
}

// ----------------------
// Устойчивый Tuner
// ----------------------
func StartTuner(sem *Sem, win *Window, tuneInterval time.Duration, lowP95, highP95 float64, minLimit, maxLimit int) {
	var smoothedP95 float64
	alpha := 0.3       // сглаживание
	hysteresis := 10.0 // мс

	go func() {
		t := time.NewTicker(tuneInterval)
		defer t.Stop()
		for range t.C {
			snap := win.Snapshot()
			p95 := percentile(snap, 0.95)
			if smoothedP95 == 0 {
				smoothedP95 = p95
			} else {
				smoothedP95 = alpha*p95 + (1-alpha)*smoothedP95
			}

			curLimit := sem.Limit()
			if smoothedP95 > highP95+hysteresis && curLimit > minLimit {
				sem.SetLimit(curLimit - 1)
			} else if smoothedP95 < lowP95-hysteresis && curLimit < maxLimit {
				sem.SetLimit(curLimit + 1)
			}
		}
	}()
}

// ----------------------
// Демонстрация
// ----------------------
func TestSem(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// параметры
	initLimit := 5
	minLimit := 2 // не даём упасть до 1
	maxLimit := 50
	windowSize := 200
	tuneInterval := time.Second
	lowP95 := 80.0   // мс
	highP95 := 200.0 // мс

	sem := NewSem(initLimit)
	win := NewWindow(windowSize)

	var processed int64
	var dropped int64

	jobRate := 120 // jobs/sec
	stopAfter := 20 * time.Second

	// Producer
	go func() {
		t := time.NewTicker(time.Second / time.Duration(jobRate))
		defer t.Stop()
		end := time.Now().Add(stopAfter)
		for time.Now().Before(end) {
			<-t.C
			if sem.TryAcquire() {
				atomic.AddInt64(&processed, 1)
				go func() {
					start := time.Now()
					// 80% быстрые, 20% длинные
					if rnd.Intn(100) < 80 {
						time.Sleep(time.Millisecond * time.Duration(30+rand.Intn(30)))
					} else {
						time.Sleep(time.Millisecond * time.Duration(200+rand.Intn(200)))
					}
					lat := float64(time.Since(start).Milliseconds())
					win.Add(lat)
					sem.Release()
				}()
			} else {
				atomic.AddInt64(&dropped, 1) // fast-fail rejected
			}
		}
	}()

	// запуск тюнера
	StartTuner(sem, win, tuneInterval, lowP95, highP95, minLimit, maxLimit)

	// Метрики
	printTicker := time.NewTicker(time.Second)
	defer printTicker.Stop()
	start := time.Now()
	for range printTicker.C {
		el := time.Since(start).Seconds()
		snap := win.Snapshot()
		p50 := percentile(snap, 0.50)
		p95 := percentile(snap, 0.95)
		p99 := percentile(snap, 0.99)
		fmt.Printf("t=%2.0fs limit=%2d cur=%2d processed=%4d dropped=%4d p50=%3.0fms p95=%3.0fms p99=%3.0fms\n",
			el, sem.Limit(), sem.Cur(),
			atomic.LoadInt64(&processed),
			atomic.LoadInt64(&dropped),
			p50, p95, p99,
		)
		if el >= stopAfter.Seconds()+1 {
			break
		}
	}
	fmt.Println("done")
}
