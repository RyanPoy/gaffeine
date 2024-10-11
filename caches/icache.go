package cache

type Cache interface {
	Get(key string) (err error)
	Set(key string, value []byte) (err error)
}
