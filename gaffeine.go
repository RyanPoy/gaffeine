package gaffeine

import (
	"gaffeine/caches"
	"gaffeine/global"
)

func NewBuilder[K global.Key]() *Gaffeine[K] {
	return &Gaffeine[K]{
		maximumSize:   -1,
		maximumWeight: -1,
	}
}

type Gaffeine[K global.Key] struct {
	maximumSize   int   // 最大cache的数量
	maximumWeight int64 // 最大权重
}

func (g *Gaffeine[K]) MaximumSize(size int) *Gaffeine[K] {
	g.maximumSize = size
	return g
}

func (g *Gaffeine[K]) MaximumWeight(weight int64) *Gaffeine[K] {
	g.maximumWeight = weight
	return g
}

func (g *Gaffeine[K]) Build() caches.Cache[K] {
	//if g.maximumWeight != -1 { // 走基于权重的设置
	//	return &caches.WeightCache[K]{}
	//}
	if g.maximumSize == -1 { // 不走基于权重的设置
		return &caches.SizeCache[K]{}
	}
	return nil
}
