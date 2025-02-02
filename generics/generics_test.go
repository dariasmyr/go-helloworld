package generics

import (
	"fmt"
	"testing"
)

type Data[T any] struct {
	data []T
}

func (d *Data[T]) GetData() []T {
	return d.data
}

func TestGenerics(t *testing.T) {
	data := Data[int]{data: []int{10, 20, 30}}
	fmt.Println("Result:", data.GetData())
	data1 := Data[string]{data: []string{"A", "B", "C"}}
	fmt.Println("Result1:", data1.GetData())
}
