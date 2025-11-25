package starlarkutil

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"
)

var ErrOutOfRange = errors.New("out of range")

func AsInt64(x starlark.Int) (int64, error) {
	val, ok := x.Int64()
	if !ok {
		return 0, fmt.Errorf("%s %w", x, ErrOutOfRange)
	}
	return val, nil
}

func AsUint64(x starlark.Int) (uint64, error) {
	val, ok := x.Uint64()
	if !ok {
		return 0, fmt.Errorf("%s %w", x, ErrOutOfRange)
	}
	return val, nil
}
