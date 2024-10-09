package frequncy_sketch_test

import (
	fs "gaffeine/frequncy_sketch"
	"gaffeine/utils"
	"github.com/stretchr/testify/assert"
	"math"
	"math/bits"
	"testing"
)

const item = 129022619

func makeSketch(maximumSize int) *fs.FrequencySketch[int] {
	var sketch = fs.New[int]()
	sketch.EnsureCapacity(maximumSize)
	return sketch
}

func TestConstruct(t *testing.T) {
	sketch := makeSketch(512)
	assert.Equal(t, 0, sketch.Size)
	assert.Equal(t, 5120, sketch.SampleSize)
	assert.Equal(t, 63, sketch.BlockMask)
	assert.Equal(t, 512, len(sketch.Table))

	sketch = makeSketch(500) // 要找500的最接近的2次幂的数作为大小，所以内部是512
	assert.Equal(t, 0, sketch.Size)
	assert.Equal(t, 5000, sketch.SampleSize)
	assert.Equal(t, 63, sketch.BlockMask)
	assert.Equal(t, 512, len(sketch.Table))
}

func TestEnsureCapacity_negative(t *testing.T) {
	sketch := makeSketch(-1)
	assert.Equal(t, 0, sketch.Size)
	assert.Equal(t, 80, sketch.SampleSize)
	assert.Equal(t, 0, sketch.BlockMask)
	assert.Equal(t, 8, len(sketch.Table))
}

func TestEnsureCapacity_smaller(t *testing.T) {
	sketch := makeSketch(512)
	size := len(sketch.Table)

	sketch.EnsureCapacity(size / 2) // 第2次重新开辟空间时候，大小比第二次的大，所以扩大
	assert.Equal(t, size, len(sketch.Table))
	assert.Equal(t, 10*size, sketch.SampleSize)
	assert.Equal(t, size>>3-1, sketch.BlockMask)
}

func TestEnsureCapacity_larger(t *testing.T) {
	sketch := makeSketch(512)
	size := len(sketch.Table)
	newSize := size * 2
	sketch.EnsureCapacity(newSize) // 第2次重新开辟空间时候，大小比第一次的小，所以维持不变
	assert.Equal(t, newSize, len(sketch.Table))
	assert.Equal(t, 10*newSize, sketch.SampleSize)
	assert.Equal(t, (newSize>>3)-1, sketch.BlockMask)
}

func TestEnsureCapacity_maximum(t *testing.T) {
	sketch := makeSketch(512)
	newSize := math.MaxInt32/10 + 1
	sketch.EnsureCapacity(newSize)
	assert.Equal(t, math.MaxInt32, sketch.SampleSize)
	assert.Equal(t, utils.CeilingPowerOfTwo32(newSize), len(sketch.Table))
	assert.Equal(t, (len(sketch.Table)>>3)-1, sketch.BlockMask)
}

func TestIncrement_once(t *testing.T) {
	sketch := makeSketch(512)
	sketch.Increment(item)
	assert.Equal(t, 1, sketch.Frequency(item))
}

func TestIncrement_max(t *testing.T) {
	sketch := makeSketch(512)
	for i := 0; i < 20; i++ {
		sketch.Increment(item)
	}
	assert.Equal(t, 15, sketch.Frequency(item))
}

func TestIncrement_distinct(t *testing.T) {
	sketch := makeSketch(512)
	sketch.Increment(item)
	sketch.Increment(item + 1)
	assert.Equal(t, 1, sketch.Frequency(item))
	assert.Equal(t, 1, sketch.Frequency(item+1))
	assert.Equal(t, 0, sketch.Frequency(item+2))
}

func TestIncrement_zero(t *testing.T) {
	sketch := makeSketch(512)
	sketch.Increment(0)
	assert.Equal(t, 1, sketch.Frequency(0))
}

func TestReset(t *testing.T) {
	reset := false
	var sketch = makeSketch(64)

	for i := 1; i < 20*len(sketch.Table); i++ {
		sketch.Increment(i)
		if sketch.Size != i {
			reset = true
			break
		}
	}
	assert.True(t, reset)
	assert.LessOrEqual(t, sketch.Size, sketch.SampleSize/2)
}

func TestFull(t *testing.T) {
	sketch := makeSketch(512)
	sketch.SampleSize = math.MaxInt32
	for i := 0; i < 100_000; i++ {
		sketch.Increment(i)
	}
	for _, item := range sketch.Table {
		assert.Equal(t, 64, bits.OnesCount64(uint64(item)))
	}

	sketch.Reset()
	for _, item := range sketch.Table {
		assert.Equal(t, fs.ResetMask, item)
	}
}

func TestHeavyHitters(t *testing.T) {
	sketch := fs.New[float64]()
	sketch.EnsureCapacity(512)
	for i := 100; i < 100_000; i++ {
		sketch.Increment(float64(i))
	}

	for i := 0; i < 10; i += 2 {
		for j := 0; j < i; j++ {
			sketch.Increment(float64(i))
		}
	}

	// A perfect popularity count yields an array [0, 0, 2, 0, 4, 0, 6, 0, 8, 0]
	popularity := make([]int, 10)
	for i := 0; i < 10; i++ {
		popularity[i] = sketch.Frequency(float64(i))
	}

	for i := 0; i < len(popularity); i++ {
		if i == 0 || i == 1 || i == 3 || i == 5 || i == 7 || i == 9 {
			assert.LessOrEqual(t, popularity[i], popularity[2])
		} else if i == 2 {
			assert.LessOrEqual(t, popularity[2], popularity[4])
		} else if i == 4 {
			assert.LessOrEqual(t, popularity[4], popularity[6])
		} else if i == 6 {
			assert.LessOrEqual(t, popularity[6], popularity[8])
		}
	}
}
