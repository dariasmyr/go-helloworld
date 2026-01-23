package filter

import (
	"fmt"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	bf := NewBloomFilter(1000, 10, 3)

	item1 := []byte("apple")
	item2 := []byte("banana")

	bf.Add(item1)

	fmt.Println("Contains 'apple':", bf.Contains(item1))
	fmt.Println("Contains 'banana':", bf.Contains(item2))
}
