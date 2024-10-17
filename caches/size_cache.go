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
	DataMap     map[K]*Element
	Window      *LRU[K]
	Probation   *LRU[K]
	Protected   *LRU[K]
	Sketch      *frequncy_sketch.FrequencySketch[K]
}

func NewSizeCache[K global.Key](size int) *SizeCache[K] {

	dataMap := make(map[K]*Element)
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
	if ele, ok := c.DataMap[key]; ok {
		ele.Value.(*global.Node[K]).Value = value
		return
	}

	addEle, windowCandidateEle := c.Window.Add(key, global.NewNode[K](key, value))
	c.DataMap[key] = addEle
	c.Sketch.Increment(key)
	if windowCandidateEle == nil { // 表示window不用淘汰
		return
	}
	if !c.Probation.IsFull() { // 表示probation不用淘汰, 可以直接把windowCandidateEle移动到probation
		c.Probation.Add(key, windowCandidateEle.Value)
		node := windowCandidateEle.Value.(*global.Node[K])
		node.InProbation()
		return
	}
	probationCandidateEle := c.Probation.Evict()
	if probationCandidateEle == nil {
		return
	}
	windowNode, probationNode := windowCandidateEle.Value.(*global.Node[K]), probationCandidateEle.Value.(*global.Node[K])
	windowFreq, probationFreq := c.Sketch.Frequency(windowNode.Key), c.Sketch.Frequency(probationNode.Key)

	if windowFreq < probationFreq { // 淘汰window
		delete(c.DataMap, windowNode.Key)
		c.Window.Evict()
	} else if windowFreq > probationFreq { // 淘汰probation，并把windowCandidateEle移动到probation
		delete(c.DataMap, windowNode.Key)
		c.Probation.Evict()
		c.Probation.Add(windowNode.Key, windowNode)
		windowNode.InProbation()
	} else if rand.Int()%2 == 0 { // 随机淘汰：如果随机数是偶数，淘汰window
		delete(c.DataMap, windowNode.Key)
		c.Window.Evict()
	} else {
		delete(c.DataMap, windowNode.Key)
		c.Probation.Evict()
		c.Probation.Add(windowNode.Key, windowNode)
		windowNode.InProbation()
	}
	return
}

func (c *SizeCache[K]) Get(key K) (interface{}, bool) {
	if ele, ok := c.DataMap[key]; ok {
		node := ele.Value.(*global.Node[K])
		c.Sketch.Increment(node.Key)
		return node.Value, true
	}
	return nil, false
}
