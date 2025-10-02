package idempotency

// Serializer defines the interface for converting a generic type T
// (which can be a success output or a failure output) to and from a byte slice
// for persistence in the idempotency store.
type Serializer[T any] interface {
	Marshal(v T) ([]byte, error)
	Unmarshal(bs []byte) (T, error)
}
