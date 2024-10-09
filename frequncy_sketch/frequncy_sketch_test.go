package frequncy_sketch_test

import (
	fs "gaffeine/frequncy_sketch"
	"github.com/stretchr/testify/assert"
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
