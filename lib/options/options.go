// Package options provides a generic option type that can be used for wrapping
// results and errors together in a single struct.
package options

import (
	"encoding/json"
)

type O[T any] struct {
	Value T
	Err   error
}

func Wrap[T any](v T, err error) O[T] {
	if err != nil {
		return FromError[T](err)
	}
	return FromValue(v)
}

func FromValue[T any](v T) O[T] {
	return O[T]{Value: v}
}

func FromError[T any](e error) O[T] {
	return O[T]{Err: e}
}

// MarshalJSON implements json.Marshaler. It returns one of the two possible
// json values:
// { "error": "error message" }
// or
// { "value": <value> }
func (o O[T]) MarshalJSON() ([]byte, error) {
	if o.Err != nil {
		return json.Marshal(errJson{Error: o.Err.Error()})
	}
	return json.Marshal(valueJson[T]{Value: o.Value})
}

type errJson struct {
	Error string `json:"error"`
}

type valueJson[T any] struct {
	Value T `json:"value"`
}
