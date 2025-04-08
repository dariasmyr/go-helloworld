package healthcheck

import (
	"fmt"
	"sync"
	"testing"
)

func merge[T any](wg *sync.WaitGroup, ch1, ch2 <-chan T) chan T {
	mergedCh := make(chan T)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for val := range ch1 {
			fmt.Println("Read from ch1", val)
			mergedCh <- val
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for val := range ch2 {
			fmt.Println("Read from ch2", val)
			mergedCh <- val
		}
	}()

	return mergedCh
}

func testMerge() {
	var wg sync.WaitGroup

	ch1 := make(chan int)
	ch2 := make(chan int)

	mergedCh := merge(&wg, ch1, ch2)

	go func() {
		defer close(ch1)
		fmt.Println("Start sending to ch1")
		for i := 0; i <= 5; i++ {
			fmt.Println("Sent to ch1", i)
			ch1 <- i
		}
	}()

	go func() {
		defer close(ch2)
		fmt.Println("Start sending to ch2")
		for i := 0; i <= 5; i++ {
			fmt.Println("Sent to ch2", i)
			ch2 <- i
		}
	}()

	go func() {
		wg.Wait()
		fmt.Println("Add goroutines has processed messages from ch1 and ch2, close OUT channel")
		close(mergedCh)
	}()

	for val := range mergedCh {
		fmt.Println("OUT Channel values", val)
	}

}

func TestMerge(t *testing.T) {
	testMerge()
}
