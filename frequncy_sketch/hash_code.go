package frequncy_sketch

type HashCoder[K Key] interface {
	HashCode(v K) uint64
}

type IntegerHashCoder[K Key] struct {
}

func (h *IntegerHashCoder[K]) HashCode(v K) uint64 {
	return uint64(v)
}
