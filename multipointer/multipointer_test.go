package multipointer

import (
	"container/heap"
	"fmt"
	"testing"
)

func TestMultipointer(t *testing.T) {
	t.Run("Test Simple Multipointer", func(t *testing.T) {
		arr1 := []int{1, 2, 4, 5}
		arr2 := []int{3, 3, 4}
		arr3 := []int{2, 3, 4, 5, 6}

		if res, ok := findCommonNumber(arr1, arr2, arr3); ok {
			fmt.Println(res)
		} else {
			fmt.Println("No common element")
		}
	})

	t.Run("Test Find allintersections", func(t *testing.T) {
		arr1 := []int{1, 2, 4, 5}
		arr2 := []int{3, 3, 4, 5}
		arr3 := []int{2, 3, 4, 5, 6}

		if res, ok := findAllCommonNumbers(arr1, arr2, arr3); ok {
			fmt.Println(res)
		} else {
			fmt.Println("No common element")
		}
	})

	t.Run("Test Find Smallest Range", func(t *testing.T) {
		arrs := [][]int{
			{4, 10, 15, 24, 26},
			{0, 9, 12, 20},
			{5, 18, 22, 30},
		}
		res := smallestRange(arrs)
		fmt.Println("Smallest Range:", res)
	})
}

func findCommonNumber(arr1, arr2, arr3 []int) (int, bool) {
	i, j, k := 0, 0, 0

	for i < len(arr1) && j < len(arr2) && k < len(arr3) {
		a := arr1[i]
		b := arr2[j]
		c := arr3[k]

		if a == b && b == c {
			return a, true
		}

		if a <= b && a <= c {
			i++
		} else if b <= a && b <= c {
			j++
		} else {
			k++
		}
	}

	return 0, false
}

func findAllCommonNumbers(arr1, arr2, arr3 []int) ([]int, bool) {
	result := []int{}

	i, j, k := 0, 0, 0

	for i < len(arr1) && j < len(arr2) && k < len(arr3) {
		a, b, c := arr1[i], arr2[j], arr3[k]

		if a == b && b == c {
			result = append(result, a)
		}

		if a <= b && a <= c {
			i++
		} else if b <= a && b <= c {
			j++
		} else {
			k++
		}
	}

	return result, true
}

type Element struct {
	val   int
	arr   int
	index int
}

type MinHeap []Element

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h MinHeap) Less(i, j int) bool { return h[i].val < h[j].val }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(Element))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	lastElement := old[len(old)-1]
	*h = old[:len(old)-1]
	return lastElement
}

func smallestRange(nums [][]int) []int {
	h := &MinHeap{}

	heap.Init(h)

	currMax := 0

	for row, arr := range nums {
		val := arr[0]
		heap.Push(h, Element{val, row, 0})
		if val > currMax {
			currMax = val
		}
	}

	bestStart, bestEnd := int(-1e9), int(1e9)

	for {
		minElement := heap.Pop(h).(Element)
		currMin := minElement.val

		if currMax-currMax < bestEnd-bestStart {
			bestStart, bestEnd = currMin, currMax
		}

		if minElement.index+1 == len(nums[minElement.arr]) {
			break
		}

		nextIndex := minElement.index + 1
		nextVal := nums[minElement.arr][nextIndex]

		heap.Push(h, Element{nextVal, minElement.arr, nextIndex})

		if nextVal > currMax {
			currMax = nextVal
		}
	}

	return []int{bestStart, bestEnd}
}
