package healthcheck

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func testWithSemaphor() {
	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://stackoverflow.com",
		"https://facebook.com",
		"http://localhost:8080",
		"http://localhost:8081",
	}

	wg := sync.WaitGroup{}
	sema := make(chan struct{}, 5)

	for _, url := range urls {
		wg.Add(1)
		sema <- struct{}{}
		go func(url string) {
			defer wg.Done()
			defer func() {
				<-sema
			}()
			res, err := http.Get(url)
			if err != nil || res.StatusCode != 200 {
				fmt.Println(url, "is not ok")
			} else {
				fmt.Println(url, "is ok")
			}
		}(url)
	}
	wg.Wait()
	close(sema)
}

func TestSemaphor(t *testing.T) {
	testWithSemaphor()
}
