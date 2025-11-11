package render_runtime

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"
)

const threadMetadataKey = "github.com/tronbyt/pixlet/runtime/$metadata"

func AttachToThread(t *starlark.Thread, m Metadata) {
	t.SetLocal(threadMetadataKey, m)
}

type Metadata struct {
	Width  int
	Height int
	Is2x   bool
}

func (m Metadata) ScaledWidth() int {
	if m.Is2x {
		return m.Width * 2
	}
	return m.Width
}

func (m Metadata) ScaledHeight() int {
	if m.Is2x {
		return m.Height * 2
	}
	return m.Height
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

		var raw bool

		if err := starlark.UnpackArgs(
			string(d),
			args, kwargs,
			"raw?", &raw,
		); err != nil {
			return nil, fmt.Errorf("unpacking arguments for %s: %w", string(d), err)
		}

		var val int
		switch d {
		case dimensionWidth:
			if raw {
				val = m.Width
			} else {
				val = m.ScaledWidth()
			}
		case dimensionHeight:
			if raw {
				val = m.Height
			} else {
				val = m.ScaledHeight()
			}
		default:
			return nil, fmt.Errorf("%w: %d", ErrUnknownDimension, d)
		}

		return starlark.MakeInt(val), nil
	}
}

func is2x(thread *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	m, ok := thread.Local(threadMetadataKey).(Metadata)
	if !ok {
		return nil, ErrNoMetadata
	}

	return starlark.Bool(m.Is2x), nil
}
