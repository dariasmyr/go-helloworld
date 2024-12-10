package order_preservation

import (
	"fmt"
	"testing"
)

func TestProcessBlocksOrderPreservationWithPool(t *testing.T) {
	numBlocks := 10
	numWorkers := 3
	inputChannel := make(chan int, numBlocks)
	outputChannel := make(chan struct {
		index int
		value int
	}, numBlocks)

	for worker := 0; worker < numWorkers; worker++ {
		go func(workerId int) {
			for block := range inputChannel {
				fmt.Printf("Worker %d processing block %d\n", workerId, block)
				result := ProcessBlock(block)
				outputChannel <- struct {
					index int
					value int
				}{
					index: block,
					value: result,
				}
			}
		}(worker)
	}

	fmt.Println("Sending blocks to input channel...")

	go func() {
		for i := 1; i <= numBlocks; i++ {
			inputChannel <- i
			fmt.Printf("Sending block %d to input channel\n", i)
		}

		close(inputChannel)
	}()

	results := make([]int, numBlocks)

	for i := 0; i < numBlocks; i++ {
		result := <-outputChannel
		fmt.Printf("Received result for block %d: %d\n", result.index, result.value)
		results[result.index-1] = result.value
	}

	close(outputChannel)

	expected := []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}

	for i, v := range results {
		if v != expected[i] {
			t.Errorf("Incorrect value at index %d: got %d, expected %d", i, v, expected[i])
		}
	}
}
