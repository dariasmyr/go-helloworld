package order_preservation

import "testing"

func ProcessBlock(block int) int {
	return block * 2
}

func TestProcessBlocksOrderPreservation(t *testing.T) {
	numBlocks := 10
	inputChannel := make(chan int, numBlocks)
	outputChannel := make(chan struct {
		index int
		value int
	}, numBlocks)

	go func() {
		for block := range inputChannel {
			outputChannel <- struct {
				index int
				value int
			}{
				index: block,
				value: ProcessBlock(block),
			}
		}

		close(outputChannel)
	}()

	go func() {
		for i := 1; i <= numBlocks; i++ {
			inputChannel <- i
		}

		close(inputChannel)
	}()

	results := make([]int, numBlocks)

	for result := range outputChannel {
		results[result.index-1] = result.value
	}

	expected := []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20}

	for i, v := range results {
		if v != expected[i] {
			t.Errorf("Incorrect value at index %d: got %d, expected %d", i, v, expected[i])
		}
	}
}
