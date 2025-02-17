package main

import "fmt"

type IntAlias = int
type StringAlias = string
type SetAlias[P comparable] = map[P]bool

type SetString = SetAlias[string]
type SetInt = SetAlias[int]

type AllowedTypes interface {
	IntAlias | StringAlias | SetInt | SetString
}

func NewBox[T AllowedTypes](value T) {
	fmt.Println("Box value:", value)
}

func main() {
	NewBox(42)
	NewBox("Hello Go")

	NewBox(map[string]bool{"a": true, "b": false})
	// NewBox(3.14)
}
