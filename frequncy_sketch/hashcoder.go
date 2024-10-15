package frequncy_sketch

import (
	"gaffeine/global"
	"math"
)

func hashcode[T global.Key](v T) uint32 {
	switch x := any(v).(type) {
	case int:
		return uint32(x)
	case uint:
		return uint32(x)
	case int8:
		return uint32(x)
	case uint8:
		return uint32(x)
	case int16:
		return uint32(x)
	case uint16:
		return uint32(x)
	case int32:
		return uint32(x)
	case uint32:
		return x
	case int64:
		return uint32(x)
	case uint64:
		return uint32(x)
	case float32:
		return math.Float32bits(x)
	case float64:
		//bits := math.Float64bits(any(v).(float64))
		bits := math.Float64bits(x)
		hash := bits ^ (bits >> 32) // 与 Java 的哈希计算相同
		return uint32(hash)
	case string:
		r := 0
		for _, b := range any(v).(string) {
			r = 31*r + int(b)
		}
		return uint32(r)
	default:
		panic("not support this type")
	}
}
