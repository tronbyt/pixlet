package colorutil

import (
	"errors"
	"fmt"
	"image/color"

	"go.starlark.net/starlark"
)

var ErrInvalid = errors.New("input is not a color or a hex string")

func Parse(v starlark.Value) (color.Color, error) {
	switch v := v.(type) {
	case starlark.String:
		c, err := ParseHex(string(v))
		if err != nil {
			return nil, fmt.Errorf("invalid hex string: %w", err)
		}
		return c, nil
	case *Color:
		return v.NRGBA, nil
	default:
		return nil, ErrInvalid
	}
}
