package global

type Key interface {
	~int | ~uint | ~float32 | ~float64
}
