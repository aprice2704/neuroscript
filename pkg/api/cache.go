package api

// Cache is an interface for a content-addressed cache.
// A full implementation will be provided later.
type Cache interface {
	Get(key [32]byte) ([]byte, bool)
	Put(key [32]byte, value []byte) error
}
