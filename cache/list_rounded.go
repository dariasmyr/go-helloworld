package cache

type node[T interface{}] struct {
	val  T
	next *node[T]
	prev *node[T]
	list *ListRounded[T]
}

type ListRounded[T interface{}] struct {
	capacity int
	root     node[T] // root is used to as the first and the last element pointed to one root node
	len      int
}

func NewListRounded[T interface{}](capacity int) *ListRounded[T] {
	l := &ListRounded[T]{capacity: capacity}
	l.root.next = &l.root //So that the first element in the root could be called by l.root.next
	l.root.prev = &l.root //So that the last element in the root could be called by l.root.prev
	return l
}

// Insert operation inserts e after at
func (l *ListRounded[T]) insert(e, at *node[T]) *node[T] {
	e.prev = at
	e.next = at.next

	e.next.prev = e
	e.prev.next = e
	e.list = l
	l.len++
	return e
}

// Remove operation removed e from its current position
func (l *ListRounded[T]) remove(e *node[T]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil
	e.prev = nil
	e.list = nil
	l.len--
}

// Move operation removed e from its current position and inserts e after at
func (l *ListRounded[T]) move(e, at *node[T]) {
	// Remove at from current position
	e.prev.next = e.next
	e.next.prev = e.prev

	// Link at to e
	e.prev = at
	e.next = at.next

	// Link e to at
	e.next.prev = e
	e.prev.next = e
}
