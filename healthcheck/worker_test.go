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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	urlChan := make(chan string)
	resChan := make(chan Response)

	urls := make([]string, 5)

	ulrsToAdd := []string{
		"https://google.com",
		"https://github.com",
		"https://stackoverflow.com",
		"https://facebook.com",
		"http://localhost:8084",
		"http://localhost:8081",
	}

	for i, _ := range urls {
		urls[i] = ulrsToAdd[i]
	}

	workersCount := 5

	go func() {
		for _, url := range urls {
			urlChan <- url
		}
		close(urlChan)
	}()

	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					fmt.Println("context cancelled, don't start the task")
				case url, ok := <-urlChan:
					if !ok {
						return
					} else {
						processUrl(ctx, url, resChan)
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for res := range resChan {
		if res.status == "200" {
			mu.Lock()
			fmt.Println("Successfully received a response 200 from url ", res.url)
			count++
			mu.Unlock()
		}
		if count == 2 {
			cancel()
			break
		}
	}
}

func processUrl(ctx context.Context, url string, resChan chan Response) {
	reqCtx, cancelReq := context.WithTimeout(ctx, 2*time.Second)
	defer cancelReq()

	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		fmt.Println("could not create request", err)
		resChan <- Response{url, "500"}
	}

	// Instead of creatinq request with http.NewRequestWithContext and executing it with http.DefaultClient.Do we can just use res, err := http.Get(url), but it will not include context
	res, err := http.DefaultClient.Do(req)
	if err != nil || res == nil {
		fmt.Println("could not send request", err)
		resChan <- Response{url, "500"}
		return
	} else {
		defer res.Body.Close()
		if res.StatusCode != 200 {
			resChan <- Response{url, "500"}
		} else {
			resChan <- Response{url, "200"}
		}
	}
}

func TestHealthCheckWithPool(t *testing.T) {
	testWithWorker()
}
