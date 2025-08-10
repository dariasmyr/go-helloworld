package sort

import (
	"fmt"
	"sync"
	"testing"
)

func concurrentQuickSort(arr []string) []string {
	if len(arr) <= 1 {
		return arr
	}
	if len(arr) < 1000 {
		return syncQuickSort(arr) // sync quicksort for small arrays
	}

	pivot := arr[len(arr)-1]
	var left, right []string
	for _, v := range arr[:len(arr)-1] {
		if v <= pivot {
			left = append(left, v)
		} else {
			right = append(right, v)
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var sortedLeft, sortedRight []string
	go func() {
		defer wg.Done()
		sortedLeft = concurrentQuickSort(left)
	}()
	go func() {
		defer wg.Done()
		sortedRight = concurrentQuickSort(right)
	}()
	wg.Wait()

	return append(append(sortedLeft, pivot), sortedRight...)
}

func syncQuickSort(arr []string) []string {
	if len(arr) <= 1 {
		return arr
	}
	fmt.Println("Initial array", arr)

	pivot := arr[len(arr)-1]
	fmt.Println("Pivot element", pivot)
	var left []string
	var right []string

	for _, value := range arr[:len(arr)-1] {
		if value <= pivot {
			left = append(left, value)
		} else {
			right = append(right, value)
		}
	}

	fmt.Println("Left", left)
	fmt.Println("Right", right)

	sortedLeft := syncQuickSort(left)
	sortedRight := syncQuickSort(right)

	return append(append(sortedLeft, pivot), sortedRight...)
}

func mergeSort(arr []string) []string {
	if len(arr) <= 1 {
		return arr
	}

	mid := len(arr) / 2

	left := mergeSort(arr[:mid])
	right := mergeSort(arr[mid:])

	return merge(left, right)
}

func merge(left, right []string) []string {
	result := make([]string, 0, len(left)+len(right))
	i, j := 0, 0

	for i < len(left) && j < len(right) {
		if left[i] < right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}

	result = append(result, left[i:]...)
	result = append(result, right[j:]...)

	return result
}

func TestQuickSort(t *testing.T) {
	arr := []string{"o", "a", "b", "g", "k"}
	fmt.Println("Arr", arr)

	qSortedArr := syncQuickSort(arr)
	fmt.Println("Quick Sort result: ", qSortedArr)

	cSortedArr := concurrentQuickSort(arr)
	fmt.Println("Concurrent quick sort result: ", cSortedArr)

	mSortedArr := mergeSort(arr)
	fmt.Println("Merge Sort result: ", mSortedArr)
}
