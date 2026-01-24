package animation

import (
	"github.com/tronbyt/gg"
)

// Rotate transforms by rotating by a given angle in degrees.
//
// DOC(Angle): Angle to rotate by in degrees.
type Rotate struct {
	Angle float64 `starlark:"angle,required"`
}

func (r Rotate) Apply(ctx *gg.Context, origin Vec2f, rounding Rounding) {
	ctx.RotateAbout(gg.Radians(r.Angle), origin.X, origin.Y)
}

func (r Rotate) Interpolate(other Transform, progress float64) (result Transform, ok bool) {
	if other, ok := other.(Rotate); ok {
		return Rotate{Lerp(r.Angle, other.Angle, progress)}, true
	}

	return RotateDefault, false
}

var RotateDefault = Rotate{0.0}
