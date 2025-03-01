package heap

import (
	"container/heap"
	"reflect"
	"testing"
)

type WordCount struct {
	word  string
	count int
}

type PriorityHeap []WordCount

func (ph PriorityHeap) Len() int {
	return len(ph)
}

func (ph PriorityHeap) Less(i, j int) bool {
	return ph[i].count < ph[j].count || (ph[i].count == ph[j].count && ph[i].word > ph[j].word)
}

func (ph PriorityHeap) Swap(i, j int) {
	ph[i], ph[j] = ph[j], ph[i]
}

func (ph *PriorityHeap) Push(x interface{}) {
	*ph = append(*ph, x.(WordCount))
}

func (ph *PriorityHeap) Pop() interface{} {
	old := *ph
	n := len(old)
	lastElement := old[n-1]
	*ph = old[:n-1]
	return lastElement
}

func topKFrequent(words []string, k int) []string {
	freq := map[string]int{}
	for _, word := range words {
		freq[word]++
	}

	ph := &PriorityHeap{}
	heap.Init(ph)

	for word, count := range freq {
		heap.Push(ph, WordCount{word, count})
		if ph.Len() > k {
			heap.Pop(ph)
		}
	}

	result := make([]string, k)
	for i := k - 1; i >= 0; i-- {
		result[i] = heap.Pop(ph).(WordCount).word
	}

	return result
}

func TestTopKFrequent(t *testing.T) {
	tests := []struct {
		words    []string
		k        int
		expected []string
	}{
		{
			words:    []string{"i", "love", "leetcode", "i", "love", "coding"},
			k:        2,
			expected: []string{"i", "love"},
		},
		{
			words:    []string{"the", "day", "is", "sunny", "the", "the", "the", "sunny", "is", "is"},
			k:        4,
			expected: []string{"the", "is", "sunny", "day"},
		},
	}

	for _, test := range tests {
		result := topKFrequent(test.words, test.k)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("For input %v and k=%d, expected %v but got %v", test.words, test.k, test.expected, result)
		}
	}
}
