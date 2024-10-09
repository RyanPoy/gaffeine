package frequncy_sketch

import (
	"math"
	"math/bits"
)

func CeilingPowerOfTwo32(x int) int {
	// From Hacker's Delight, Chapter 3, Harry S. Warren Jr.
	tmp := -1 * bits.LeadingZeros32(uint32(x-1))
	if tmp < 0 {
		tmp = 32 + tmp
	}
	return 1 << tmp
}

func CeilingPowerOfTwo64(x int64) int64 {
	// From Hacker's Delight, Chapter 3, Harry S. Warren Jr.
	var n int64 = 1
	tmp := -1 * bits.LeadingZeros64(uint64(x-1))
	if tmp < 0 {
		tmp = 64 + tmp
	}
	return n << tmp
}

type Key interface {
	~int
}

type FrequencySketch[K Key] struct {
	KeyType    K
	SampleSize int // 需要进行Reset的容量
	BlockMask  int // 一个块(8个int64大小）的掩码
	Size       int // 当前已经使用的计数器个数，这个是一个评估值，不是一个精确值

	Table     []int64
	HashCoder HashCoder[K]
}

func New[K Key]() *FrequencySketch[K] {
	sketch := FrequencySketch[K]{
		Table:      nil,
		SampleSize: 0,
		BlockMask:  0,
		Size:       0,
	}
	switch any(sketch.KeyType).(type) {
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		sketch.HashCoder = &IntegerHashCoder[K]{}
	default:
		panic("unsupported type")
	}

	return &sketch
}

func (f *FrequencySketch[K]) EnsureCapacity(maximumSize int) *FrequencySketch[K] {
	if maximumSize < 0 {
		maximumSize = 8
	}

	maximum := int(math.Min(float64(maximumSize), math.MaxInt>>1))
	if len(f.Table) >= maximum {
		return f
	}
	newSize := int(math.Max(float64(CeilingPowerOfTwo32(maximum)), 8))
	f.Table = make([]int64, newSize)
	if maximumSize == 0 {
		f.SampleSize = 10
	} else {
		f.SampleSize = 10 * maximum
	}

	// a）>>3，是因为：64位架构CPU的一个缓存块大小是64个字节。
	// 				 而8个int64为一个块，刚好是64个字节，从而有更快的读取速度
	// b）-1，是因为：len(f.Table)>>3得到的数一定是一个首位是1，其他位是0的数。
	// 				-1后，首位是0，其他位是1，从而得到一个掩码。
	f.BlockMask = int(uint(len(f.Table))>>3 - 1)

	if f.SampleSize <= 0 {
		f.SampleSize = math.MaxInt
	}
	f.Size = 0

	return f
}

func (f *FrequencySketch[K]) Increment(key K) *FrequencySketch[K] {
	return f
}

func (f *FrequencySketch[K]) Frequency(key K) int {
	return 0
}
