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

func Min[T1, T2 Key](a T1, b T2) float64 {
	return math.Min(float64(a), float64(b))
}

func Max[T1, T2 Key](a T1, b T2) float64 {
	return math.Max(float64(a), float64(b))
}
