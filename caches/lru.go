package caches

import (
	"gaffeine/global"
)

type LRU[K global.Key] struct {
	data map[K]*Element
	size int
	lst  *List
}

func NewLRU[K global.Key](size int, data map[K]*Element) *LRU[K] {
	return &LRU[K]{
		data: data,
		lst:  NewList(),
		size: size,
	}
}

func (lru *LRU[K]) Len() int             { return lru.lst.Len() }
func (lru *LRU[K]) Size() int            { return lru.size }
func (lru *LRU[K]) IsFull() bool         { return lru.lst.Len() >= lru.size }
func (lru *LRU[K]) WhetherToEvict() bool { return lru.lst.Len() > lru.size }

// Add adds a new key-value pair to the LRU.
// it returns the add element and next eviction element.
// if it is not full, the eviction element is nil.
func (lru *LRU[K]) Add(key K, value any) (*Element, *Element) {
	if ele, ok := lru.data[key]; ok {
		lru.lst.MoveToFront(ele)
		return ele, nil
	}
	ele := lru.lst.PushFront(value)
	if lru.WhetherToEvict() {
		return ele, lru.lst.Back()
	}
	return ele, nil
}

// Evict removes the least recently used element from the LRU.
func (lru *LRU[K]) Evict() *Element {
	if !lru.WhetherToEvict() {
		return nil
	}
	back := lru.lst.Back()
	lru.lst.Remove(back)
	return back
}
