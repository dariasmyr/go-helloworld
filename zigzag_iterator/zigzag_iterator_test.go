package zigzag_iterator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ZigzagIterator struct {
	data [][]int // All vectors list
	ids  []int   // Current vector value index in each vector
	pos  int     // Current vector number
}

func NewZigzagIterator(vectors ...[]int) *ZigzagIterator {
	return &ZigzagIterator{
		data: vectors,
		ids:  make([]int, len(vectors)),
	}
}

func (z *ZigzagIterator) HasNext() bool {
	start := z.pos
	// Check if current vector value id does not exceed vector's length (z.ids[z.pos] - current value index in current vector)
	for z.ids[z.pos] == len(z.data[z.pos]) {
		// Vector has ho elements left, calc next vector
		z.pos = (z.pos + 1) % len(z.data)
		fmt.Println("Next vector", z.pos)
		if z.pos == start {
			fmt.Println("Checked: all vectors has no elements left")
			return false
		}
	}
	// There are elements left if current vector, return true
	fmt.Println("There are elements left in current vector: ", z.pos)
	return true
}

func (z *ZigzagIterator) Next() (int, error) {
	fmt.Println("Current vector: ", z.pos, "; current value id:", z.ids[z.pos])

	// Check if current vector value id does not exceed vector's length
	if z.ids[z.pos] < len(z.data[z.pos]) {
		valueIdx := z.ids[z.pos]
		val := z.data[z.pos][valueIdx]
		fmt.Println("Calculated value: ", val)
		z.ids[z.pos] = valueIdx + 1
		fmt.Println("Incrementing value index: ", valueIdx)
		z.pos = (z.pos + 1) % len(z.data)
		fmt.Println("Change vector position to: ", z.pos)
		return val, nil
	} else {
		return 0, fmt.Errorf("next position index exceeds length on the vector")
	}
}

func TestZigzagIterator(t *testing.T) {
	t.Run("MixedLengths", func(t *testing.T) {
		v1 := []int{1, 2}
		v2 := []int{3, 4, 5, 6}
		v3 := []int{7, 8, 9, 10}

		iter := NewZigzagIterator(v1, v2, v3)

		var result []int
		for iter.HasNext() {
			val, err := iter.Next()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			result = append(result, val)
		}

		expected := []int{1, 3, 7, 2, 4, 8, 5, 9, 6, 10}
		assert.Equal(t, expected, result)
	})

	t.Run("DecreasingLengths", func(t *testing.T) {
		v1 := []int{1, 2, 3, 4}
		v2 := []int{5, 6, 7}
		v3 := []int{8, 9}

		iter := NewZigzagIterator(v1, v2, v3)

		var result []int
		for iter.HasNext() {
			val, err := iter.Next()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			result = append(result, val)
		}

		expected := []int{1, 5, 8, 2, 6, 9, 3, 7, 4}
		assert.Equal(t, expected, result)
	})
}
