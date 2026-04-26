package starlarkutil

import (
	"iter"

	"go.starlark.net/starlark"
)

// Enumerate returns an iter.Seq2 for the elements of the iterable value.
func Enumerate(it starlark.Iterable) iter.Seq2[int, starlark.Value] {
	return func(yield func(int, starlark.Value) bool) {
		var i int
		starlark.Elements(it)(func(v starlark.Value) bool {
			ok := yield(i, v)
			i++
			return ok
		})
	}
}
