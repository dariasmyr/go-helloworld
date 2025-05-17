package thumbnail_generator

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// === RateLimiter ===

type RateLimiter struct {
	mu           sync.Mutex
	tokens       float64
	maxTokens    float64
	tokensPerSec float64
	lastRefill   time.Time
}

func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:       float64(burst),
		maxTokens:    float64(burst),
		tokensPerSec: rate,
		lastRefill:   time.Now(),
	}
}

func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	newTokens := elapsed * rl.tokensPerSec
	if newTokens > 0 {
		rl.tokens += newTokens
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.refill()
	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

// === Job Pipeline ===

type Job struct {
	Conn net.Conn
	Data []byte
}

type Result struct {
	Conn net.Conn
	Data []byte
}

// === Thumbnail mock ===

func generateThumbnail(data []byte) []byte {
	return append([]byte("[thumb]"), data...)
}

// === Connection Handler ===

func handleConnection(ctx context.Context, conn net.Conn, jobs chan<- Job) {
	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Println("conn close error:", err)
		}
	}()
	reader := bufio.NewReader(conn)

	for {
		err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)) // avoid hang
		if err != nil {
			fmt.Println("error setting read deadline:", err)
			return
		}

		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("client closed connection:", err)
			} else {
				fmt.Println("read error:", err)
			}
			return
		}

		select {
		case jobs <- Job{Conn: conn, Data: data}:
			// continue
		case <-ctx.Done():
			return
		}
	}
}

// === Worker ===

func worker(ctx context.Context, id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			fmt.Printf("[worker %d] processing\n", id)
			res := generateThumbnail(job.Data)
			results <- Result{Conn: job.Conn, Data: res}
		case <-ctx.Done():
			return
		}
	}
}

// === Writer Goroutine ===

func resultWriter(results <-chan Result, done chan struct{}) {
	for res := range results {
		_, err := res.Conn.Write(res.Data)
		if err != nil {
			fmt.Println("write error:", err)
		}
	}
	close(done)
}

// === Server ===

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-stop
		cancel()
		signal.Stop(stop)
	}()

	listener, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}
	defer func(listener net.Listener) {
		err = listener.Close()
		if err != nil {
			panic(err)
		}
	}(listener)
	fmt.Println("Listening on :9090")

	jobs := make(chan Job, 100)
	results := make(chan Result, 100)
	done := make(chan struct{})

	go resultWriter(results, done)

	limiter := NewRateLimiter(3, 10)

	go func() {
		<-ctx.Done()
		closeErr := listener.Close()
		if closeErr != nil {
			return
		}
	}()

	var jobProducer sync.WaitGroup
	go func() {
	acceptLoop:
		for {
			conn, acceptErr := listener.Accept()
			if acceptErr != nil {
				select {
				case <-ctx.Done():
					break acceptLoop
				default:
					fmt.Println("accept error:", acceptErr)
					continue
				}
			}

			if !limiter.Allow() {
				n, limitErr := fmt.Fprintln(conn, "Too many connections. Try later.")
				if limitErr != nil {
					fmt.Printf("error writing refusal message: %v, bytes written: %d\n", limitErr, n)
				}
				closeConnErr := conn.Close()
				if closeConnErr != nil {
					fmt.Println("close error:", closeConnErr)
					return
				}
				continue
			}
			jobProducer.Add(1)
			go func() {
				defer jobProducer.Done()
				handleConnection(ctx, conn, jobs)
			}()
		}
	}()

	go func() {
		jobProducer.Wait()
		close(jobs)
	}()

	var workerWg sync.WaitGroup
	numWorkers := 8
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go worker(ctx, i, jobs, results, &workerWg)
	}

	workerWg.Wait()
	close(results)

	// Wait for resultWriter to finish his job
	<-done
}
