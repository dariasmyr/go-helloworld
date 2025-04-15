package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"testing"
)

func doSomething() (result int, err error) {
	result = 0
	f, err := os.Open("phuong-secrets.txt")
	// defer func is called after return valuea are valuated and before the func has been closed
	// pass by closure - defer get pointer to result and err values - so we can change the return values
	defer func() {
		err = errors.Join(err, f.Close())
		result = 5
		fmt.Println("Defer with closure", result, err)
	}()
	// pass by args values - values are copied in the moment of defining defer func - re cant change the return values
	defer func(result int, err error) {
		err = errors.Join(err, f.Close())
		fmt.Println("Defer with value args", result, err)
	}(result, err)
	result = 2
	if err != nil {
		return result, err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return 1, nil
}

func Test(t *testing.T) {
	result, err := doSomething()
	fmt.Println("Func result", result, err)
}
