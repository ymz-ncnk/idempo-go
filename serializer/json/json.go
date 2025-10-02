package json

import "encoding/json"

type JSONSerializer[T any] struct{}

func (s JSONSerializer[T]) Marshal(v T) (bs []byte, err error) {
	return json.Marshal(v)
}

func (s JSONSerializer[T]) Unmarshal(bs []byte) (v T, err error) {
	err = json.Unmarshal(bs, &v)
	return
}
