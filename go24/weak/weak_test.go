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

func TestBasic(t *testing.T) {
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

func TestGC(t *testing.T) {
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
	mu        sync.Mutex
	items     map[int]weak.Pointer[CacheData[T]]
	itemsWeak map[weak.Pointer[CacheData[T]]]*CacheData[T]
}

func (c *Cache[T]) Add(id int, data *CacheData[T]) {
	weakPointer := weak.Make(data)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[id] = weakPointer
}

func (c *Cache[T]) AddWeakKey(data *CacheData[T]) weak.Pointer[CacheData[T]] {
	weakPointer := weak.Make(data)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.itemsWeak[weakPointer] = data
	return weakPointer
}

func (c *Cache[T]) Get(id int) *CacheData[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	if weakPoiner, exists := c.items[id]; exists {
		return weakPoiner.Value() // Return strong pointer (*Data)
	}
	return nil
}

func (c *Cache[T]) GetCacheByWeakPointer(weakPtr weak.Pointer[CacheData[T]]) *CacheData[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	if data, exists := c.itemsWeak[weakPtr]; exists {
		return data // Return *CacheData pointer
	}
	return nil
}

type Node struct {
	Name string
	Next *Node
}

func TestCache(t *testing.T) {
	t.Run("Test Cache", func(t *testing.T) {
		cache := &Cache[string]{items: make(map[int]weak.Pointer[CacheData[string]])}

		data1 := &CacheData[string]{ID: 1, Value: "Object 1"}

		cache.Add(1, data1)

		fmt.Printf("Cache 1 before GC: %+v\n", cache.Get(1))

		// Call GC manually
		runtime.GC()

		fmt.Printf("Cache after GC: %+v\n", cache.Get(1))
	})

	t.Run("Test Cache Map", func(t *testing.T) {
		cache := &Cache[string]{itemsWeak: make(map[weak.Pointer[CacheData[string]]]*CacheData[string])}

		data1 := &CacheData[string]{ID: 1, Value: "Object 1"}

		weakKey := cache.AddWeakKey(data1)

		fmt.Printf("Weak Pointer to Cache 1 before GC: %+v\n", weakKey)
		fmt.Printf("Cache 1 before GC: %+v\n", cache.GetCacheByWeakPointer(weakKey))

		runtime.GC()

		fmt.Printf("weakKey after GC: %+v\n", weakKey)
		fmt.Printf("Cache 1 after GC: %+v\n", cache.GetCacheByWeakPointer(weakKey))
	})

	t.Run("Test Weak Circular Pointers", func(t *testing.T) {
		node1 := &Node{Name: "Node 1"}
		node2 := &Node{Name: "Node 2"}

		// Circular pointer: node1 refer to node2 and vice versa
		node1.Next = node2
		node2.Next = node1

		weakNode1 := weak.Make(&node1)
		weakNode2 := weak.Make(&node2)

		// Setting links to weak pointers
		// At the same time, the garbage collector can delete node1 and node2 when there are no more strong references to them.
		fmt.Printf("Before GC, weakNode1: %+v, weakNode2: %+v\n", weakNode1.Value(), weakNode2.Value())

		runtime.GC()

		// After garbage collection, we see that references to node1 and node2 objects can be released,
		// since they are now controlled by a weak pointer rather than a strong reference
		if weakNode1.Value() == nil {
			fmt.Println("weakNode1 was collected by GC")
		} else {
			fmt.Println("weakNode1 is still alive")
		}

		if weakNode2.Value() == nil {
			fmt.Println("weakNode2 was collected by GC")
		} else {
			fmt.Println("weakNode2 is still alive")
		}
	})
}
