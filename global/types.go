package global

const (
	WindowPos = iota
	ProbationPos
	ProtectedPos
)

type Key interface {
	int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | float32 | float64 | string
}

type Node[K Key] struct {
	Key   K
	Value any
	pos   int
}

func NewNode[K Key](v any) *Node[K] {
	return &Node[K]{Value: v, pos: WindowPos}
}
func (v *Node[K]) InWindow()    { v.pos = WindowPos }
func (v *Node[K]) InProbation() { v.pos = ProbationPos }
func (v *Node[K]) InProtected() { v.pos = ProtectedPos }
func (v *Node[K]) isInWindow() bool {
	return v.pos == WindowPos
}
func (v *Node[K]) isInProbation() bool {
	return v.pos == ProbationPos
}
func (v *Node[K]) isInProtected() bool {
	return v.pos == ProtectedPos
}
