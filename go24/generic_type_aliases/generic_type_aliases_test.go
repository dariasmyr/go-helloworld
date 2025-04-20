package go24

import (
	"fmt"
	"testing"
)

type StringAlias = string
type StringArrAlias = []string

type SetAlias[P comparable] = map[P]bool
type SetString = SetAlias[string]
type SetInt = SetAlias[int]

type TypeValue interface {
	int | string
}

type A[P comparable] = TypeValue
type B = A[bool]

type AllowedTypes interface {
	StringAlias | StringArrAlias | SetInt | SetString | B
}

func NewBox[T AllowedTypes](value T) {
	fmt.Println("Box value:", value)
}

func TestAliases(t *testing.T) {
	t.Run("Test pipeline success", func(t *testing.T) {
		NewBox("Hello Go")

		NewBox([]string{"a", "b"})

		NewBox(3)

		//NewBox(true) не пройдет, так как не bool does not satisfy AllowedTypes
	})
}
