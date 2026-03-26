package color

import (
	"fmt"
	"image/color"
	"sync"

	"github.com/tronbyt/pixlet/internal/colorutil"
	"github.com/tronbyt/pixlet/starlarkutil"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const ModuleName = "color"

var (
	once   sync.Once
	module starlark.StringDict
)

func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		module = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"rgb": starlark.NewBuiltin("rgb", parseRGB),
					"hex": starlark.NewBuiltin("hex", parseHex),
					"hsv": starlark.NewBuiltin("hsv", parseHSV),
				},
			},
		}
		module.Freeze()
	})

	return module, nil
}

func parseRGB(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var rParam, gParam, bParam starlark.Int
	aParam := starlark.Value(starlark.None)

	if err := starlark.UnpackArgs(
		fn.Name(), args, kwargs,
		"r", &rParam,
		"g", &gParam,
		"b", &bParam,
		"a?", &aParam,
	); err != nil {
		return nil, err
	}

	r, err := starlarkutil.AsInt[int64](rParam)
	if err != nil {
		return nil, fmt.Errorf("parsing r: %w", err)
	}

	g, err := starlarkutil.AsInt[int64](gParam)
	if err != nil {
		return nil, fmt.Errorf("parsing g: %w", err)
	}

	b, err := starlarkutil.AsInt[int64](bParam)
	if err != nil {
		return nil, fmt.Errorf("parsing b: %w", err)
	}

	var a int64
	switch aParam := aParam.(type) {
	case starlark.Int:
		if a, err = starlarkutil.AsInt[int64](aParam); err != nil {
			return nil, fmt.Errorf("parsing a: %w", err)
		}
	case starlark.NoneType:
		a = 255
	default:
		return nil, fmt.Errorf("a must be an int, got %T", aParam)
	}

	c := color.NRGBA{
		R: uint8(max(0, min(r, 255))),
		G: uint8(max(0, min(g, 255))),
		B: uint8(max(0, min(b, 255))),
		A: uint8(max(0, min(a, 255))),
	}

	return &colorutil.Color{NRGBA: c}, nil
}

func parseHex(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var hex starlark.String

	if err := starlark.UnpackArgs(
		fn.Name(), args, kwargs,
		"value", &hex,
	); err != nil {
		return nil, err
	}

	if len(hex) == 0 {
		return &colorutil.Color{}, nil
	}

	c, err := colorutil.ParseHex(string(hex))
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %w", err)
	}

	return &colorutil.Color{NRGBA: c}, nil
}

func parseHSV(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var hParam, sParam, vParam starlark.Float
	aParam := starlark.Value(starlark.None)

	if err := starlark.UnpackArgs(
		fn.Name(), args, kwargs,
		"h", &hParam,
		"s", &sParam,
		"v", &vParam,
		"a?", &aParam,
	); err != nil {
		return nil, err
	}

	h := float64(hParam)
	s := float64(sParam)
	v := float64(vParam)

	var a int64
	switch aParam := aParam.(type) {
	case starlark.Int:
		var err error
		if a, err = starlarkutil.AsInt[int64](aParam); err != nil {
			return nil, fmt.Errorf("parsing a: %w", err)
		}
		a = max(0, min(a, 255))
	case starlark.NoneType:
		a = 255
	default:
		return nil, fmt.Errorf("a must be an int, got %T", aParam)
	}

	c := colorutil.ParseHSV(h, s, v, uint8(a))
	return &colorutil.Color{NRGBA: c}, nil
}
