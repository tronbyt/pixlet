package animation

import (
	"github.com/tronbyt/gg"
)

// Transform by shearing by a given X and Y angle in degrees.
//
// DOC(XAngle): The angle to shear horizontally in degrees.
// DOC(YAngle): The angle to shear vertically in degrees.
type Shear struct {
	XAngle float64 `starlark:"x_angle"`
	YAngle float64 `starlark:"y_angle"`
}

func (self Shear) Apply(ctx *gg.Context, origin Vec2f, rounding Rounding) {
	ctx.ShearAbout(gg.Radians(self.XAngle), gg.Radians(self.YAngle), origin.X, origin.Y)
}

func (self Shear) Interpolate(other Transform, progress float64) (result Transform, ok bool) {
	if other, ok := other.(Shear); ok {
		return Shear{
			XAngle: Lerp(self.XAngle, other.XAngle, progress),
			YAngle: Lerp(self.YAngle, other.YAngle, progress),
		}, true
	}

	return ShearDefault, false
}

var ShearDefault = Shear{}
