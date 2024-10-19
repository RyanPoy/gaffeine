package caches

import (
	"gaffeine/frequncy_sketch"
	"gaffeine/global"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type SizeCache[K global.Key] struct {
	MaximumSize int
	DataMap     map[K]*Element[K]
	Window      *LRU[K]
	Probation   *LRU[K]
	Protected   *LRU[K]
	Sketch      *frequncy_sketch.FrequencySketch[K]
}

func NewSizeCache[K global.Key](size int) *SizeCache[K] {

	dataMap := make(map[K]*Element[K])
	windowSize := int(float32(size) * 0.02)
	probationSize := int(float32(size) * 0.2)
	protectedSize := probationSize * 4
	if windowSize <= 0 {
		windowSize = 2
	}
	if probationSize <= 0 {
		probationSize = 2
	}
	if protectedSize <= 0 {
		protectedSize = 8
	}
	maxSize := windowSize + probationSize + protectedSize

	return &SizeCache[K]{
		MaximumSize: maxSize,
		DataMap:     dataMap,
		Window:      NewLRU(windowSize, dataMap),
		Probation:   NewLRU(probationSize, dataMap),
		Protected:   NewLRU(protectedSize, dataMap),
		Sketch:      frequncy_sketch.New[K]().EnsureCapacity(maxSize),
	}
}

// Set sets key and value to cache.
// step:
// hashmap.put(node)
// 如果node的权重大于windowq的最大权重，push到windowq的first，否则push到windowq的last
// 如果window的当前权重大于window最大权重，挪动window的first，放到probation的last，直到window的当前权重小于等于window的最大权重。到此：window的当前权重已经收缩到合理值了。
// loop：如果cache的当前权重超出最大权重，进行淘汰：
//
//	如果probation的 victim(first) 和 candidate(last) 进行对比，按照FrequencyCandidate 和 FrequencyVictim 和 随机数 一起来判断淘汰 Victim 或者 Candidate。到此：Cache的当前权重已经收缩到合理值了。
func (c *SizeCache[K]) Set(key K, value interface{}) {
	if ele, ok := c.DataMap[key]; ok { // 表示key已经存在，更新value
		ele.Value = value
		return
	}

	ele := c.Window.PushFront(value)
	ele.Key = key

	c.DataMap[key] = ele
	c.Sketch.Increment(key)

	windowCandidateEle, ok := c.evictFromLRU(c.Window, func() bool { return c.Window.NeedEvict() })
	if !ok {
		return
	}

	if !c.Probation.IsFull() {
		// 如果probation没有满，则把windowCandidateEle移动到probation的first
		c.Probation.InsertAtFront(windowCandidateEle)
		windowCandidateEle.InProbation()
		return
	}

	// 到这里，就证明probation也可能需要淘汰。所以需要进行选举了
	probationCandidateEle := c.Probation.Back()
	if probationCandidateEle == nil {
		return
	}

	windowFreq := c.Sketch.Frequency(windowCandidateEle.Key)
	probationFreq := c.Sketch.Frequency(probationCandidateEle.Key)

	if windowFreq < probationFreq { // 直接淘汰window
		delete(c.DataMap, windowCandidateEle.Key)
	} else if windowFreq > probationFreq { // 淘汰probation，并把windowCandidateEle移动到probation
		delete(c.DataMap, probationCandidateEle.Key)
		c.Probation.EvictBack()
		c.Probation.InsertAtFront(windowCandidateEle)
		windowCandidateEle.InProbation()
	} else if rand.Int()%2 == 0 { // 随机淘汰：如果随机数是偶数，淘汰window
		delete(c.DataMap, windowCandidateEle.Key)
	} else {
		delete(c.DataMap, probationCandidateEle.Key)
		c.Probation.EvictBack()
		c.Probation.InsertAtFront(windowCandidateEle)
		windowCandidateEle.InProbation()
	}
	return
}

func (c *SizeCache[K]) evictFromLRU(lru *LRU[K], checkFunc func() bool) (*Element[K], bool) {
	if !checkFunc() {
		return nil, false
	}
	ele := lru.EvictBack()
	for {
		if !lru.NeedEvict() {
			break
		}
		ele2 := lru.EvictBack()
		if c.Sketch.Frequency(ele2.Key) > c.Sketch.Frequency(ele.Key) {
			ele = ele2
		}
	}
	return ele, true
}
func (c *SizeCache[K]) Get(key K) (interface{}, bool) {
	if ele, ok := c.DataMap[key]; ok {
		c.Sketch.Increment(ele.Key)
		return ele.Value, true
	}
	return nil, false
}
