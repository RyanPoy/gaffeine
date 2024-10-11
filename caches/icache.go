package caches

import "gaffeine/global"

type Cache[K global.Key] interface {
	Get(key K) (interface{}, error)
	Set(key K, value interface{}) error
}
