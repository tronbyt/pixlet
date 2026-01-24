package animation

import (
	"github.com/tronbyt/gg"
)

// Shear transforms by shearing by a given X and Y angle in degrees.
//
// DOC(XAngle): The angle to shear horizontally in degrees.
// DOC(YAngle): The angle to shear vertically in degrees.
type Shear struct {
	XAngle float64 `starlark:"x_angle"`
	YAngle float64 `starlark:"y_angle"`
}

func (s Shear) Apply(ctx *gg.Context, origin Vec2f, rounding Rounding) {
	ctx.ShearAbout(gg.Radians(s.XAngle), gg.Radians(s.YAngle), origin.X, origin.Y)
}

func (s Shear) Interpolate(other Transform, progress float64) (result Transform, ok bool) {
	if other, ok := other.(Shear); ok {
		return Shear{
			XAngle: Lerp(s.XAngle, other.XAngle, progress),
			YAngle: Lerp(s.YAngle, other.YAngle, progress),
		}, true
	}

	return ShearDefault, false
}

var ShearDefault = Shear{}
