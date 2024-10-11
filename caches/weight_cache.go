package caches

import "gaffeine/global"

type WeightCache[K global.Key] struct{}

func (c *WeightCache[K]) Get(key K) (interface{}, error) {
	return nil, nil
}
func (c *WeightCache[K]) Set(key K, value interface{}) error {
	return nil
}
