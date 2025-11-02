package render_runtime

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"
)

type Metadata struct {
	Width  int
	Height int
}

const threadMetadataKey = "tidbyt.dev/pixlet/runtime/$metadata"

func AttachToThread(t *starlark.Thread, m Metadata) {
	t.SetLocal(threadMetadataKey, m)
}

type dimensionType uint8

const (
	dimensionWidth dimensionType = iota
	dimensionHeight
)

var (
	ErrUnknownDimension = errors.New("unknown dimension")
	ErrNoMetadata       = errors.New("no metadata available")
)

func dimension(d dimensionType) func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		m, ok := thread.Local(threadMetadataKey).(Metadata)
		if !ok {
			return nil, ErrNoMetadata
		}

		var val int
		switch d {
		case dimensionWidth:
			val = m.Width
		case dimensionHeight:
			val = m.Height
		default:
			return nil, fmt.Errorf("%w: %d", ErrUnknownDimension, d)
		}

		return starlark.MakeInt(val), nil
	}
}
