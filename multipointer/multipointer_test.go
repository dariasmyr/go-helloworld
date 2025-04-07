package multipointer

import (
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
}

func findCommonNumber(arr1, arr2, arr3 []int) (int, bool) {
	i, j, k := 0, 0, 0

	for i < len(arr1) && j < len(arr2) && k < len(arr3) {
		a, b, c := arr1[i], arr2[j], arr3[k]
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
