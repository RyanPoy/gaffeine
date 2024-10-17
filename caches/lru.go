package caches

import "gaffeine/global"

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lru implements a doubly linked lru.
//
// To iterate over a lru (where l is a *LRU):
//
//	for e := l.Front(); e != nil; e = e.Next() {
//		// do something with e.Value
//	}
const (
	WindowPos = iota
	ProbationPos
	ProtectedPos
)

// Element is an element of a linked lru.
type Element[K global.Key] struct {
	next, prev *Element[K]
	Key        K
	Value      any // The value stored with this element.
	pos        int
}

func NewElement[K global.Key](key K, v any) *Element[K] {
	return &Element[K]{
		Key:   key,
		Value: v,
		pos:   WindowPos,
	}
}

func (e *Element[K]) InWindow()    { e.pos = WindowPos }
func (e *Element[K]) InProbation() { e.pos = ProbationPos }
func (e *Element[K]) InProtected() { e.pos = ProtectedPos }
func (e *Element[K]) IsInWindow() bool {
	return e.pos == WindowPos
}
func (e *Element[K]) IsInProbation() bool {
	return e.pos == ProbationPos
}
func (e *Element[K]) IsInProtected() bool {
	return e.pos == ProtectedPos
}

// LRU represents a doubly linked lru.
// The zero value for LRU is an empty lru ready to use.
type LRU[K global.Key] struct {
	root Element[K] // sentinel lru element, only &root, root.prev, and root.next are used
	len  int        // current lru length excluding (this) sentinel element
	size int
	data map[K]*Element[K]
}

// Init initializes or clears lru.
func (l *LRU[K]) Init() *LRU[K] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// New returns an initialized lru.
func NewLRU[K global.Key](size int, data map[K]*Element[K]) *LRU[K] {
	lst := new(LRU[K]).Init()
	lst.data = data
	lst.size = size
	return lst
}

// Len returns the number of elements of lru l.
// The complexity is O(1).
func (l *LRU[K]) Len() int        { return l.len }
func (l *LRU[K]) Size() int       { return l.size }
func (l *LRU[K]) IsFull() bool    { return l.Len() >= l.size }
func (l *LRU[K]) NeedEvict() bool { return l.Len() > l.size }

// Add adds a new key-value pair to the LRU.
// Front returns the first element of lru l or nil if the lru is empty.
func (l *LRU[K]) Front() *Element[K] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of lru l or nil if the lru is empty.
func (l *LRU[K]) Back() *Element[K] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// insert inserts e after at, increments l.len, and returns e.
func (l *LRU[K]) insert(e, at *Element[K]) *Element[K] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	l.len++
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at).
func (l *LRU[K]) insertValue(v any, at *Element[K]) *Element[K] {
	return l.insert(&Element[K]{Value: v}, at)
}

// remove removes e from its lru, decrements l.len
func (l *LRU[K]) remove(e *Element[K]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	l.len--
}

// move moves e to next to at.
func (l *LRU[K]) move(e, at *Element[K]) {
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

// Remove removes e from l if e is an element of lru l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *LRU[K]) Remove(e *Element[K]) any {
	l.remove(e)
	return e.Value
}

// PushFront inserts a new element e with value v at the front of lru l and returns e.
func (l *LRU[K]) PushFront(v any) *Element[K] { return l.insertValue(v, &l.root) }

// PushBack inserts a new element e with value v at the back of lru l and returns e.
func (l *LRU[K]) PushBack(v any) *Element[K] { return l.insertValue(v, l.root.prev) }

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the lru is not modified.
// The mark must not be nil.
func (l *LRU[K]) InsertBefore(v any, mark *Element[K]) *Element[K] {
	// see comment in LRU.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the lru is not modified.
// The mark must not be nil.
func (l *LRU[K]) InsertAfter(v any, mark *Element[K]) *Element[K] {
	// see comment in LRU.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of lru l.
// If e is not an element of l, the lru is not modified.
// The element must not be nil.
func (l *LRU[K]) MoveToFront(e *Element[K]) {
	if l.root.next == e {
		return
	}
	// see comment in LRU.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of lru l.
// If e is not an element of l, the lru is not modified.
// The element must not be nil.
func (l *LRU[K]) MoveToBack(e *Element[K]) {
	if l.root.prev == e {
		return
	}
	// see comment in LRU.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the lru is not modified.
// The element and mark must not be nil.
func (l *LRU[K]) MoveBefore(e, mark *Element[K]) {
	if e == mark {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the lru is not modified.
// The element and mark must not be nil.
func (l *LRU[K]) MoveAfter(e, mark *Element[K]) {
	if e == mark {
		return
	}
	l.move(e, mark)
}

// it returns the add element and next eviction element.
// if it is not full, the eviction element is nil.
//func (l *LRU[K]) Add(key K, value any) (*Element[K], *Element[K]) {
//	if ele, ok := l.data[key]; ok {
//		l.MoveToFront(ele)
//		return ele, nil
//	}
//	ele := l.PushFront(value)
//	if l.NeedEvict() {
//		return ele, l.Back()
//	}
//	return ele, nil
//}

// Evict removes the least recently used element from the LRU.
func (l *LRU[K]) EvictBack() *Element[K] {
	if !l.NeedEvict() {
		return nil
	}
	back := l.Back()
	l.Remove(back)
	return back
}
