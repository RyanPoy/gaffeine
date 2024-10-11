package utils

import (
	"container/list"
)

var aaaa_data list.List

type Element struct {
	Value interface{}
	next  *Element
	prev  *Element
	list  *ConcurrentList
}

func (e *Element) Next() *Element {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev returns the previous list element or nil.
func (e *Element) Prev() *Element {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

type ConcurrentList struct {
	root Element
	len  int
}

func NewConcurrentList() *ConcurrentList {
	return &ConcurrentList{
		root: Element{},
		len:  0,
	}
}

func (l *ConcurrentList) Len() int {
	return l.len
}

// Front returns the first element of list l or nil if the list is empty.
func (l *ConcurrentList) Front() *Element {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *ConcurrentList) Back() *Element {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// insert inserts e after at, increments l.len, and returns e.
func (l *ConcurrentList) insert(e, at *Element) *Element {
	e.prev = at
	e.next = at.next

	e.prev.next = e
	e.next.prev = e

	e.list = l
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *ConcurrentList) insertValue(v any, at *Element) *Element {
	return l.insert(&Element{Value: v}, at)
}

// remove removes e from its list, decrements l.len
func (l *ConcurrentList) remove(e *Element) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks

	e.list = nil
	l.len--
}

// move moves e to next to at.
func (l *ConcurrentList) move(e, at *Element) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *ConcurrentList) Remove(e *Element) any {
	if e.list == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(e)
	}
	return e.Value
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *ConcurrentList) PushFront(v any) *Element {
	return l.insertValue(v, &l.root)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *ConcurrentList) PushBack(v any) *Element {
	return l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *ConcurrentList) InsertBefore(v any, mark *Element) *Element {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *ConcurrentList) InsertAfter(v any, mark *Element) *Element {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *ConcurrentList) MoveToFront(e *Element) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *ConcurrentList) MoveToBack(e *Element) {
	if e.list != l || l.root.prev == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *ConcurrentList) MoveBefore(e, mark *Element) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *ConcurrentList) MoveAfter(e, mark *Element) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *ConcurrentList) PushBackList(other *ConcurrentList) {
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *ConcurrentList) PushFrontList(other *ConcurrentList) {
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}

//// Add adds a new node to the list back
//func (l *ConcurrentList) Add(value interface{}) error {
//	ele := &Element{Value: value}
//	for {
//
//		tail := lst.tail
//		prevTail := tail.prev
//
//		// Try to insert the new node before the tail
//		ele.prev = prevTail
//		ele.next = tail
//
//		if atomic.CompareAndSwapPointer(
//			(*unsafe.Pointer)(unsafe.Pointer(&prevTail.next)),
//			unsafe.Pointer(tail),
//			unsafe.Pointer(ele),
//		) {
//			// Update tail's previous pointer
//			atomic.CompareAndSwapPointer(
//				(*unsafe.Pointer)(unsafe.Pointer(&tail.prev)),
//				unsafe.Pointer(prevTail),
//				unsafe.Pointer(ele),
//			)
//			return
//		}
//
//		// Tail is outdated; move tail to the next node
//		atomic.CompareAndSwapPointer(
//			(*unsafe.Pointer)(unsafe.Pointer(&l.tail)),
//			unsafe.Pointer(tail),
//			unsafe.Pointer(prevTail.next),
//		)
//	}
//}
