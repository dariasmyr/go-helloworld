package main

import (
	"fmt"
	"runtime"
	"weak"
)

type Data struct {
	ID   int
	Name string
}

func testBasic() {
	var x int = 42
	weakPointer := weak.Make(&x)
	var d Data
	weakPointer2 := weak.Make(&d)

	fmt.Println("weakPointer", weakPointer.Value())
	if st := weakPointer.Value(); st != &x {
		fmt.Printf("weak pointer is not the same as strong pointer: %p vs. %p", st, &x)
	}
	fmt.Println("weakPointer2", weakPointer2.Value())
}

func testGC() {
	x := &Data{ID: 1, Name: "Dasha"}

	weakPointer := weak.Make(x)

	fmt.Printf("Before GC: weakPointer.Value() = %+v\n", weakPointer.Value())

	if st := weakPointer.Value(); st != x {
		fmt.Printf("weak pointer is not the same as strong pointer: %p vs. %p\n", st, x)
	} else {
		fmt.Printf("weak pointer value is the same as strong pointer: %p vs. %p\n", st, x)
	}

	//Call GC manually, x poiner is released as there is ni usage of it in the code
	runtime.GC()

	//Check Weak pointer after GC
	if weakPointer.Value() == nil {
		fmt.Println("After GC: weakPointer.Value() = nil (object collected)")
	} else {
		fmt.Printf("After GC: weakPointer.Value() = %+v (object still alive)\n", weakPointer.Value())
	}

}

func main() {
	testBasic()
	testGC()
}
