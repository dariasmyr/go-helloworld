package main

import "fmt"

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

	fmt.Println("Recursive call for left", left)
	sortedLeft := quickSort(left)
	fmt.Println("Sorted left", sortedLeft)

	fmt.Println("Recursive call for right", right)
	sortedRight := quickSort(right)
	fmt.Println("Sorted right", sortedRight)

	return append(append(sortedLeft, pivot), sortedRight...)
}

func main() {
	arr := []string{"o", "a", "b", "g", "k"}
	fmt.Println("Arr", arr)

	sortedArr := quickSort(arr)
	fmt.Println("sortedArr", sortedArr)
}
