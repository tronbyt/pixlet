package colorutil

import (
	"fmt"
	"image/color"

	"go.starlark.net/starlark"
)

type Color struct {
	color.NRGBA
}

const (
	AttrR    = "r"
	AttrG    = "g"
	AttrB    = "b"
	AttrA    = "a"
	AttrH    = "h"
	AttrS    = "s"
	AttrV    = "v"
	AttrHex  = "hex"
	AttrRGB  = "rgb"
	AttrRGBA = "rgba"
	AttrHSV  = "hsv"
	AttrHSVA = "hsva"
)

func (c *Color) AttrNames() []string {
	return []string{
		AttrR, AttrG, AttrB, AttrA,
		AttrH, AttrS, AttrV,
		AttrHex, AttrRGB, AttrRGBA, AttrHSV, AttrHSVA,
	}
}

func (c *Color) Attr(name string) (starlark.Value, error) {
	switch name {
	case AttrR:
		return starlark.MakeInt(int(c.R)), nil
	case AttrG:
		return starlark.MakeInt(int(c.G)), nil
	case AttrB:
		return starlark.MakeInt(int(c.B)), nil
	case AttrA:
		return starlark.MakeInt(int(c.A)), nil
	case AttrH:
		h, _, _, _ := FormatHSV(c.NRGBA)
		return starlark.Float(h), nil
	case AttrS:
		_, s, _, _ := FormatHSV(c.NRGBA)
		return starlark.Float(s), nil
	case AttrV:
		_, _, v, _ := FormatHSV(c.NRGBA)
		return starlark.Float(v), nil
	case AttrHex:
		return starlark.NewBuiltin(name, c.getHex), nil
	case AttrRGB:
		return starlark.NewBuiltin(name, c.getRGB), nil
	case AttrRGBA:
		return starlark.NewBuiltin(name, c.getRGBA), nil
	case AttrHSV:
		return starlark.NewBuiltin(name, c.getHSV), nil
	case AttrHSVA:
		return starlark.NewBuiltin(name, c.getHSVA), nil
	default:
		return nil, nil
	}
}

func (c *Color) SetField(name string, val starlark.Value) error {
	switch name {
	case AttrR, AttrG, AttrB, AttrA:
		i, err := starlark.AsInt32(val)
		if err != nil {
			return fmt.Errorf("value for %q must be an integer", name)
		}
		i = max(0, min(i, 255))

		switch name {
		case AttrR:
			c.R = uint8(i)
		case AttrG:
			c.G = uint8(i)
		case AttrB:
			c.B = uint8(i)
		case AttrA:
			c.A = uint8(i)
		}
	case AttrH, AttrS, AttrV:
		f, ok := starlark.AsFloat(val)
		if !ok {
			return fmt.Errorf("value for %q must be a float", name)
		}

		h, s, v, a := FormatHSV(c.NRGBA)
		switch name {
		case AttrH:
			h = f
		case AttrS:
			s = f
		case AttrV:
			v = f
		}

		c.NRGBA = ParseHSV(h, s, v, a)
	default:
		return starlark.NoSuchAttrError(fmt.Sprintf("cannot assign to field %q", name))
	}
	return nil
}

func (c *Color) String() string {
	return `Color("` + FormatHex(c.NRGBA) + `")`
}
func (c *Color) Type() string         { return "Color" }
func (c *Color) Freeze()              {}
func (c *Color) Truth() starlark.Bool { return true }

func (c *Color) Hash() (uint32, error) {
	hash := uint32(c.R)<<24 | uint32(c.G)<<16 | uint32(c.B)<<8 | uint32(c.A)
	return hash, nil
}

func (c *Color) getHex(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	return starlark.String(FormatHex(c.NRGBA)), nil
}

func (c *Color) getRGB(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	return starlark.Tuple{
		starlark.MakeInt(int(c.R)),
		starlark.MakeInt(int(c.G)),
		starlark.MakeInt(int(c.B)),
	}, nil
}

func (c *Color) getRGBA(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	return starlark.Tuple{
		starlark.MakeInt(int(c.R)),
		starlark.MakeInt(int(c.G)),
		starlark.MakeInt(int(c.B)),
		starlark.MakeInt(int(c.A)),
	}, nil
}

func (c *Color) getHSV(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	h, s, v, _ := FormatHSV(c.NRGBA)
	return starlark.Tuple{
		starlark.Float(h),
		starlark.Float(s),
		starlark.Float(v),
	}, nil
}

func (c *Color) getHSVA(_ *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	h, s, v, _ := FormatHSV(c.NRGBA)
	return starlark.Tuple{
		starlark.Float(h),
		starlark.Float(s),
		starlark.Float(v),
		starlark.MakeInt(int(c.A)),
	}, nil
}
