package worker_pool

import (
	"fmt"
	"sync"
)

func simpleWorker(tasks <-chan string, results chan<- map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()
	localResult := make(map[string]int)
	for task := range tasks {
		processed := len(task)
		localResult[task] = processed
	}
	results <- localResult
}

func main() {
	tasks := make(chan string, 10)
	results := make(chan map[string]int, 5)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go simpleWorker(tasks, results, &wg)
	}

	go func() {
		for _, task := range []string{"task1", "task2", "task3"} {
			tasks <- task
		}
		close(tasks)
	}()

	wg.Wait()
	close(results)

	finalResult := make(map[string]int)
	for res := range results {
		for k, v := range res {
			finalResult[k] = v
		}
	}

	fmt.Println(finalResult)
}
