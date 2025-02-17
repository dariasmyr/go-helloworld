package go24

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"weak"
)

type Data struct {
	ID   int
	Name string
}

func testBasic(t *testing.T) {
	t.Run("Test Basic", func(t *testing.T) {
		var x int = 42
		weakPointer := weak.Make(&x)
		var d Data
		weakPointer2 := weak.Make(&d)

		fmt.Println("weakPointer", weakPointer.Value())
		if st := weakPointer.Value(); st != &x {
			fmt.Printf("weak pointer is not the same as strong pointer: %p vs. %p", st, &x)
		}
		fmt.Println("weakPointer2", weakPointer2.Value())
	})
}

func testGC(t *testing.T) {
	t.Run("Test GC", func(t *testing.T) {
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

	})
}

// Cache test case
type CacheValue interface {
	string | int
}

type CacheData[T CacheValue] struct {
	ID    int
	Value T
}

type Cache[T CacheValue] struct {
	mu    sync.Mutex
	items map[int]weak.Pointer[CacheData[T]]
}

func (c *Cache[T]) Add(id int, data *CacheData[T]) {
	weakPointer := weak.Make(data)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[id] = weakPointer
}

func (c *Cache[T]) Get(id int) *CacheData[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	if weakPoiner, exists := c.items[id]; exists {
		return weakPoiner.Value() // Return strong pointer (*Data)
	}
	return nil
}

func testCache(t *testing.T) {
	t.Run("Test Cache", func(t *testing.T) {
		cache := &Cache[string]{items: make(map[int]weak.Pointer[CacheData[string]])}

		data1 := &CacheData[string]{ID: 1, Value: "Object 1"}

		cache.Add(1, data1)

		fmt.Printf("Cache 1 before GC: %+v\n", cache.Get(1))

		// Initally i wanted to create a cache map where i could store both string and int values (CacheValue common type interface), but as it is restricted for methosa (in my exampe Add and Set) to have type parameters (as i wanted leave type Cache strunf without [T Cache Value] and make Add[T CacheValue] and Set[T CacheValue]). Unf it is allowed only for structures (in my example type Cache struct) to have type params. So i can create onle cache map for strings ONLY or ints ONLY)
		// data2 := &CacheData[int]{ID: 2, Value: 123}

		// cache.Add(2, data2)

		// fmt.Printf("Cache 2 before GC: %+v\n", cache.Get(2))

		// Call GC manually
		runtime.GC()

		fmt.Printf("Cache after GC: %+v\n", cache.Get(1))
		// fmt.Printf("Cache after GC: %+v\n", cache.Get(2))
	})
}
