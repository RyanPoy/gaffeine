package caches

import "gaffeine/global"

type Cache[K global.Key] interface {
	Get(key K) (interface{}, bool)
	Set(key K, value interface{})
}
