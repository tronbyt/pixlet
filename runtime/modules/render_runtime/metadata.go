package render_runtime

import (
	"errors"
	"fmt"

	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/metadata"
	"go.starlark.net/starlark"
)

type dimensionType uint8

const (
	dimensionWidth dimensionType = iota
	dimensionHeight
)

var ErrUnknownDimension = errors.New("unknown dimension")

func dimension(d dimensionType) func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		m, err := metadata.FromThread(thread)
		if err != nil {
			return nil, err
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
	m, err := metadata.FromThread(thread)
	if err != nil {
		return nil, err
	}

	return starlark.Bool(m.Is2x), nil
}
