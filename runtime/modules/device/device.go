package device

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName        = "device"
	threadMetadataKey = "tidbyt.dev/pixlet/runtime/$metadata"
)

func AttachToThread(t *starlark.Thread, m Metadata) {
	t.SetLocal(threadMetadataKey, m)
}

func LoadModule() (starlark.StringDict, error) {
	return starlark.StringDict{
		ModuleName: &starlarkstruct.Module{
			Name: ModuleName,
			Members: starlark.StringDict{
				"width":  starlark.NewBuiltin("width", dimension(dimensionWidth)),
				"height": starlark.NewBuiltin("height", dimension(dimensionHeight)),
				"is2x":   starlark.NewBuiltin("is2x", is2x),
			},
		},
	}, nil
}

type dimensionType uint8

const (
	dimensionWidth dimensionType = iota
	dimensionHeight
)

var (
	ErrUnknownDimension = fmt.Errorf("unknown dimension")
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
			"width",
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
