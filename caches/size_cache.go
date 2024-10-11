package caches

import (
	"container/list"
	"gaffeine/global"
)

type SizeCache[K global.Key] struct {
	MaximumSize   int
	WindowSize    int
	ProbationSize int
	ProtectedSize int

	Window    list.List
	Probation list.List
	Protected list.List
}

func (c *SizeCache[K]) Get(key K) (interface{}, error) {
	return nil, nil
}
func (c *SizeCache[K]) Set(key K, value interface{}) error {
	return nil
}
