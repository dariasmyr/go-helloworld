package iterators

import (
	"iter"
	"testing"
)

func Filter(seq iter.Seq[int], by func(int) bool) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := range seq {
			if by(i) {
				if !yield(i) {
					return
				}
			}
		}
	}
}

func Range(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := range n {
			if !yield(i) {
				return
			}
		}
	}
}

func Multiply(seq iter.Seq[int], n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := range seq {
			result := i * n
			if !yield(result) {
				return
			}
		}
	}
}

func TestPipeline(t *testing.T) {
	t.Run("Test pipeline success", func(t *testing.T) {
		elements := Range(100000000)
		elements = Multiply(elements, 2)
		elements = Filter(elements, func(i int) bool {
			return i%3 == 0
		})

		for i := range elements {
			println(i)
		}

	})
}
