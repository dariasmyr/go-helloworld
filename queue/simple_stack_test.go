package queue

import (
	"errors"
	"fmt"
	"testing"
)

func TestStack(t *testing.T) {
	runStack()
}

func runStack() {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("recovered from panic", r)
		}
	}()
	stack := []int{}

	//append
	stack = append(stack, 1)
	stack = append(stack, 2)
	stack = append(stack, 3)

	for len(stack) > 0 {
		// print top
		fmt.Printf("top element: %v\n", stack[len(stack)-1])

		// pop
		removedElem, err := pop(&stack)
		if err != nil {
			fmt.Println("could not pop", removedElem, err)
		}
		fmt.Printf("removed %v\n", removedElem)
	}

	errLastElem, err := pop(&stack)
	if err != nil {
		fmt.Println("could not pop", errLastElem, err)
	}
	fmt.Printf("last element: %v\n", errLastElem)
}

func pop(stack *[]int) (int, error) {
	oldStack := *stack
	if len(oldStack) == 0 {
		err := errors.New("get from empty stack")
		return 0, err
	}
	top := oldStack[len(oldStack)-1]
	*stack = oldStack[:len(oldStack)-1]
	return top, nil
}
