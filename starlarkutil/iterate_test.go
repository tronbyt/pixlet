package starlarkutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

func TestEnumerate(t *testing.T) {
	values := make([]starlark.Value, 10)
	for i := range values {
		values[i] = starlark.MakeInt(i)
	}
	list := starlark.NewList(values)

	for i, v := range Enumerate(list) {
		var j int
		require.NoError(t, starlark.AsInt(v, &j))
		assert.Equal(t, i, j)
	}
}
