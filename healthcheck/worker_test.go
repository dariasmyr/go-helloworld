package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"
)

type Response struct {
	url    string
	status string
}

var count int

func testWithWorker() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	numWorkers := 2
	var urlChan = make(chan string)
	var resChan = make(chan Response)
	urls := []string{
		"http://localhost:8084",
		"http://localhost:8081",
		"https://google.com",
		"https://github.com",
		"https://stackoverflow.com",
		"https://facebook.com",
	}

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for _, url := range urls {
			urlChan <- url
		}
		close(urlChan)
	}()

	for worker := 0; worker < numWorkers; worker++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case url, ok := <-urlChan:
					if !ok {
						return
					}
					processRequest(ctx, url, resChan)
				}
			}
		}(worker)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for res := range resChan {
		if res.status == "200" {
			mu.Lock()
			count++
			mu.Unlock()
			fmt.Println(res.url, "is ok")
		}
		if count == 2 {
			cancel()
			break
		}
	}
}

func processRequest(ctx context.Context, url string, resChan chan<- Response) {
	resCtx, cancelReq := context.WithTimeout(ctx, 5*time.Second)
	defer cancelReq()
	select {
	case <-resCtx.Done():
		fmt.Println("Response context cancelled")
		resChan <- Response{url: url, status: "500"}
	default:
		res, err := http.Get(url)
		if err != nil || res.StatusCode != 200 {
			fmt.Println(url, "is not ok")
			resChan <- Response{url: url, status: "500"}
		} else {
			resChan <- Response{url: url, status: "200"}
		}
	}
}

func TestHealthCheckWithPool(t *testing.T) {
	testWithWorker()
}
