package caches

import "gaffeine/global"

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package list implements a doubly linked list.
//
// To iterate over a list (where l is a *List):
//
//	for e := l.Front(); e != nil; e = e.Next() {
//		// do something with e.Value
//	}
const (
	WindowPos = iota
	ProbationPos
	ProtectedPos
)

type Node[K global.Key] struct {
	Key   K
	Value any
	pos   int
}

func NewNode[K global.Key](key K, v any) *Node[K] {
	return &Node[K]{
		Key:   key,
		Value: v,
		pos:   WindowPos,
	}
}

func (v *Node[K]) InWindow()    { v.pos = WindowPos }
func (v *Node[K]) InProbation() { v.pos = ProbationPos }
func (v *Node[K]) InProtected() { v.pos = ProtectedPos }
func (v *Node[K]) IsInWindow() bool {
	return v.pos == WindowPos
}
func (v *Node[K]) IsInProbation() bool {
	return v.pos == ProbationPos
}
func (v *Node[K]) IsInProtected() bool {
	return v.pos == ProtectedPos
}

// Element is an element of a linked list.
type Element[K global.Key] struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Element[K]

	// The value stored with this element.
	Value any
}

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
type List[K global.Key] struct {
	root Element[K] // sentinel list element, only &root, root.prev, and root.next are used
	len  int        // current list length excluding (this) sentinel element
}

// Init initializes or clears list l.
func (l *List[K]) Init() *List[K] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// New returns an initialized list.
func NewList[K global.Key]() *List[K] { return new(List[K]).Init() }

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *List[K]) Len() int { return l.len }

// Front returns the first element of list l or nil if the list is empty.
func (l *List[K]) Front() *Element[K] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List[K]) Back() *Element[K] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// insert inserts e after at, increments l.len, and returns e.
func (l *List[K]) insert(e, at *Element[K]) *Element[K] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *List[K]) insertValue(v any, at *Element[K]) *Element[K] {
	return l.insert(&Element[K]{Value: v}, at)
}

// remove removes e from its list, decrements l.len
func (l *List[K]) remove(e *Element[K]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	l.len--
}

// move moves e to next to at.
func (l *List[K]) move(e, at *Element[K]) {
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
func (l *List[K]) Remove(e *Element[K]) any {
	l.remove(e)
	return e.Value
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *List[K]) PushFront(v any) *Element[K] { return l.insertValue(v, &l.root) }

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *List[K]) PushBack(v any) *Element[K] { return l.insertValue(v, l.root.prev) }

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[K]) InsertBefore(v any, mark *Element[K]) *Element[K] {
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[K]) InsertAfter(v any, mark *Element[K]) *Element[K] {
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[K]) MoveToFront(e *Element[K]) {
	if l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[K]) MoveToBack(e *Element[K]) {
	if l.root.prev == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[K]) MoveBefore(e, mark *Element[K]) {
	if e == mark {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[K]) MoveAfter(e, mark *Element[K]) {
	if e == mark {
		return
	}
	l.move(e, mark)
}
