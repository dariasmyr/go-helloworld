package sort

import (
	"fmt"
	"testing"
)

func quickSort(arr []string) []string {
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

	sortedLeft := quickSort(left)
	sortedRight := quickSort(right)

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

	qSortedArr := quickSort(arr)
	fmt.Println("Quick Sort result: ", qSortedArr)

	mSortedArr := mergeSort(arr)
	fmt.Println("Merge Sort result: ", mSortedArr)
}
