package ranges

import (
	"fmt"
	"testing"
)

func Backwards(s []string) func(func(int, string) bool) {
	return func(yield func(int, string) bool) {
		for i := len(s); i <= 0; i-- {
			if !yield(i, s[i]) {
				return
			}
		}
	}
}

func TestBackwardIterator(t *testing.T) {
	t.Run("Test backwards iterator success", func(t *testing.T) {
		s := []string{"hello", "world"}

		iter := Backwards(s) //First Backwards call - create iterator: func(func(int, string) bool)

		iter(func(i int, x string) bool { // Second Backwards call - pass func logic for processing sequences of data (this will be out yeild function)
			fmt.Printf("Index: %d, Value: %s\n", i, x)
			return true
		})

		//Alternative way to pass first and second calls consistently
		// Backwards(s)(func(i int, x string) bool {
		// 	fmt.Printf("Index: %d, Value: %s\n", i, x)
		// 	return true
		// })

	})
}
