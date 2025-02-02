package generics

import (
	"fmt"
	"reflect"
	"testing"
)

type data1 struct {
	Data []int
}

type data2 struct {
	Data []string
}

type data3 struct {
	Data []byte
}

func getDataWithReflectOnly(data interface{}) interface{} {
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Struct {
		dataField := val.FieldByName("Data")
		if dataField.IsValid() && dataField.Kind() == reflect.Slice {
			return dataField.Interface()
		}
	}
	return nil
}

func getDataWithReflectAndGenerics[T any](data interface{}) []T {
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Struct {
		dataField := val.FieldByName("Data")
		if dataField.IsValid() && dataField.Kind() == reflect.Slice {
			return dataField.Interface().([]T)
		}
	}
	return nil
}

// Works only for struct with one field Data of type []T
func getDataWithGenericsOnly[T any](data struct{ Data []T }) []T {
	return data.Data
}

func TestReflect(t *testing.T) {

	data1 := data1{Data: []int{1, 2, 3}}
	fmt.Println("data1:", getDataWithReflectOnly(data1))
	fmt.Println("data1:", getDataWithReflectAndGenerics[int](data1))
	fmt.Println("data1:", getDataWithGenericsOnly(data1))

	data2 := data2{Data: []string{"A", "B", "C"}}
	fmt.Println("data2:", getDataWithReflectOnly(data2))
	fmt.Println("data2:", getDataWithReflectAndGenerics[string](data2))
	fmt.Println("data2:", getDataWithGenericsOnly(data2))

	data3 := data3{Data: []byte{1, 2, 3}}
	fmt.Println("data3:", getDataWithReflectOnly(data3))
	fmt.Println("data3:", getDataWithReflectAndGenerics[byte](data3))
	fmt.Println("data3:", getDataWithGenericsOnly(data3))

}
