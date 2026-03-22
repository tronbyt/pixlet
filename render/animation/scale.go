package animation

import (
	"github.com/tronbyt/gg"
)

// Scale transforms by scaling by a given factor.
type Scale struct {
	// Horizontal scale factor
	X float64 `starlark:"x,required"`
	// Vertical scale factor
	Y float64 `starlark:"y,required"`
}

func (s Scale) Apply(ctx *gg.Context, origin Vec2f, rounding Rounding) {
	ctx.ScaleAbout(s.X, s.Y, origin.X, origin.Y)
}

func (s Scale) Interpolate(other Transform, progress float64) (result Transform, ok bool) {
	if other, ok := other.(Scale); ok {
		return Scale{
			X: Lerp(s.X, other.X, progress),
			Y: Lerp(s.Y, other.Y, progress),
		}, true
	}

	return ScaleDefault, false
}

var ScaleDefault = Scale{X: 1.0, Y: 1.0}
