package race

import (
	"fmt"
	"sync"
	"testing"
)

func TestRaceCondition(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Hello")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("World")
	}()
	wg.Wait()
}

func increment(i *int) {
	for range 1000 {
		*i++
	}
}
func TestDataRace(t *testing.T) {
	var wg sync.WaitGroup

	var count int

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			increment(&count)
		}()
	}

	wg.Wait()
	fmt.Println(count)
}

func TestDeadlock(t *testing.T) {
	var wg sync.WaitGroup

	var mu sync.Mutex
	var mu2 sync.Mutex

	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		fmt.Println("Goroutine 1 locks mu")
		defer mu.Unlock()

		mu2.Lock()
		fmt.Println("Goroutine 1 locks mu2")
		defer mu2.Unlock()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		mu2.Lock()
		fmt.Println("Goroutine 2 locks mu2")
		defer mu2.Unlock()

		mu.Lock()
		fmt.Println("Goroutine 2 locks mu")
		defer mu.Unlock()
	}()

	wg.Wait()
}
