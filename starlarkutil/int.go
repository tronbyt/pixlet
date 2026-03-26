package starlarkutil

import (
	"go.starlark.net/starlark"
)

type asIntTypes interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func AsInt[T asIntTypes](x starlark.Int) (T, error) {
	var val T
	err := starlark.AsInt(x, &val)
	return val, err
}
