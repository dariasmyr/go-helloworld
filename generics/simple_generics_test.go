package generics

import (
	"fmt"
	"testing"
)

type TypeValue interface {
	int | string
}

type Data[T TypeValue] struct {
	data []T
}

func (d *Data[T]) GetData() []T {
	return d.data
}

func GetDataFunc[T TypeValue](data []T) []T {
	return data
}

func TestGenerics(t *testing.T) {
	data := Data[int]{data: []int{10, 20, 30}}
	fmt.Println("Result:", data.GetData())
	data1 := Data[string]{data: []string{"A", "B", "C"}}
	fmt.Println("Result1:", data1.GetData())
	data2 := []string{"D", "E", "F"}
	fmt.Println("Result3:", GetDataFunc(data2))
}
