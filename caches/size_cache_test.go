package caches_test

import (
	"gaffeine/caches"
	"github.com/stretchr/testify/assert"
	"testing"
)

func makeSizeCache(size int) *caches.SizeCache[string] {
	return caches.NewSizeCache[string](size)
}
func TestConstruct_lessSize(t *testing.T) {
	cache := makeSizeCache(4)
	assert.Equal(t, 12, cache.MaximumSize)
	assert.Equal(t, 2, cache.Window.Size())
	assert.Equal(t, 2, cache.Probation.Size())
	assert.Equal(t, 8, cache.Protected.Size())

	assert.Equal(t, 0, cache.Window.Len())
	assert.Equal(t, 0, len(cache.DataMap))
}

func TestConstruct_normal(t *testing.T) {
	cache := makeSizeCache(20)
	assert.Equal(t, 22, cache.MaximumSize)
	assert.Equal(t, 2, cache.Window.Size())
	assert.Equal(t, 4, cache.Probation.Size())
	assert.Equal(t, 16, cache.Protected.Size())
}

func TestSet_new(t *testing.T) {
	cache := makeSizeCache(4)

	k, v := "key", 10
	cache.Set(k, v)
	assert.Equal(t, 1, cache.Window.Len())

	ele, ok := cache.DataMap[k]
	assert.True(t, ok)
	assert.Equal(t, v, ele.Value.(int))
	assert.True(t, ele.IsInWindow())
}

func TestSet_update(t *testing.T) {
	cache := makeSizeCache(4)

	k, v1, v2 := "key", 10, 20
	cache.Set(k, v1)
	cache.Set(k, v2)
	assert.Equal(t, 1, cache.Window.Len())

	ele, ok := cache.DataMap[k]
	assert.True(t, ok)
	assert.Equal(t, v2, ele.Value.(int))
	assert.True(t, ele.IsInWindow())
}

func TestSet_moveToProbationFromWindowWhileProbationFrequencyMoreThanWindow(t *testing.T) {
	cache := makeSizeCache(4)
	k1, v1 := "key1", 10
	k2, v2 := "key2", 20
	k3, v3 := "key3", 30
	cache.Set(k1, v1)
	cache.Set(k2, v2)
	cache.Set(k3, v3)

	ele, _ := cache.DataMap[k1]
	assert.True(t, ele.IsInProbation())

	ele, _ = cache.DataMap[k2]
	assert.Equal(t, v2, ele.Value.(int))
	assert.True(t, ele.IsInWindow())

	ele, _ = cache.DataMap[k3]
	assert.Equal(t, v3, ele.Value.(int))
	assert.True(t, ele.IsInWindow())
}

//
//// 3、测试key、value，需要淘汰时候window，在DataMap中能看到key、value，同时旧数据淘汰了，window的容量依然合规。
//func TestSet_evictFromWindow(t *testing.T) {
//	cache := makeSizeCache(4)
//
//	// make window and probation full
//	k1, v1 := "key1", 10
//	k2, v2 := "key2", 20
//	k3, v3 := "key3", 30
//	k4, v4 := "key4", 40
//	cache.Set(k1, v1) // window: k1
//	cache.Set(k2, v2) // window: k2, k1
//	cache.Set(k3, v3) // window: k3, k2;  probation: k1
//	cache.Set(k4, v4) // window: k4, k3;  probation: k2, k1
//
//	// increase the frequency of probation elements
//	cache.Get(k1)
//	cache.Get(k2)
//
//	// will evict element k4 of window after set element k3
//	// why not k1 of probation? because k1's frequency is more than k4
//	k5, v5 := "key5", 50
//	cache.Set(k5, v5) // window: k5, k4; probation: k2, k1
//
//	_, ok := cache.Get(k3)
//	assert.False(t, ok)
//}

// 4、测试key、value，需要淘汰时候probation，在DataMap中能看到key、value，同时旧数据淘汰了，probation的容量依然合规。
// 5、

func TestGet_foundAndIncrementFrequency(t *testing.T) {
	cache := makeSizeCache(4)
	key := "key"
	cache.Set(key, 10)
	assert.Equal(t, 1, cache.Sketch.Frequency(key))

	v, ok := cache.Get(key)
	assert.True(t, ok)
	assert.Equal(t, 10, v.(int))
	assert.Equal(t, 2, cache.Sketch.Frequency(key))

}
