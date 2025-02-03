package generics

import (
	"fmt"
	"testing"
)

type data11 struct {
	Data []int
	Name string
}

type data22 struct {
	Data    []string
	Address string
}

type data33 struct {
	Data  []byte
	Phone string
}

type hasData[T any] interface {
	GetData() []T
}

func (d data11) GetData() []int {
	return d.Data
}

func (d data22) GetData() []string {
	return d.Data
}

func (d data33) GetData() []byte {
	return d.Data
}

func getData[T any, D hasData[T]](data D) []T {
	return data.GetData()
}

func TestGeneric(t *testing.T) {

	data1 := data11{Data: []int{1, 2, 3}, Name: "John"}
	// We Should implement GetData() in data11, otherwise it will panic:
	// Cannot use data11 as the type hasData[T] Type does not implement 'hasData[T]' as some methods are missing: GetData() []int
	fmt.Println("data11:", getData[int](data1))

	data2 := data22{Data: []string{"A", "B", "C"}, Address: "123 Main St"}
	fmt.Println("data22:", getData[string](data2))

	data3 := data33{Data: []byte{1, 2, 3}}
	fmt.Println("data33:", getData[byte](data3))

}
