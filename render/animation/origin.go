package animation

import (
	"image"
)

// Origin is a relative anchor point to use for scaling and rotation transforms.
//
// DOC(X): Horizontal anchor point
// DOC(Y): Vertical anchor point.
type Origin struct {
	X Percentage `starlark:"x,required"`
	Y Percentage `starlark:"y,required"`
}

func (o Origin) Transform(bounds image.Rectangle) Vec2f {
	return Vec2f{
		o.X.Value * float64(bounds.Dx()),
		o.Y.Value * float64(bounds.Dy()),
	}
}

var DefaultOrigin = Origin{
	X: Percentage{0.5},
	Y: Percentage{0.5},
}
