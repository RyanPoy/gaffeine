package frequncy_sketch

import (
	"gaffeine/global"
	"math"
)

type HashCoder[K global.Key] interface {
	HashCode(v K) uint32
}

type IntegerHashCoder[K global.Key] struct{}

func (h *IntegerHashCoder[K]) HashCode(v K) uint32 {
	return uint32(v)
}

type Float32HashCoder[K global.Key] struct{}

func (h *Float32HashCoder[K]) HashCode(v K) uint32 {
	return math.Float32bits(float32(v))
}

type Float64HashCoder[K global.Key] struct{}

func (h *Float64HashCoder[K]) HashCode(v K) uint32 {
	bits := math.Float64bits(float64(v))
	hash := bits ^ (bits >> 32) // 与 Java 的哈希计算相同
	return uint32(hash)
}

type StringHashCoder[K global.Key] struct{}

func (h *StringHashCoder[K]) HashCode(v K) uint32 {
	r := 0
	for _, b := range any(v).(string) {
		r = 31*r + int(b)
	}
	return uint32(r)
}
