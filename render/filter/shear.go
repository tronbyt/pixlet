package filter

import (
	"image"
	"math"

	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Shear shears the child widget horizontally and/or vertically.
//
// DOC(Widget): The widget to shear.
// DOC(XAngle): The angle to shear horizontally in degrees.
// DOC(YAngle): The angle to shear vertically in degrees.
//
// EXAMPLE BEGIN
//
//	filter.Shear(
//	    child = render.Image(src="...", width=64, height=64),
//	    x_angle = 10.0,
//	)
//
// EXAMPLE END
type Shear struct {
	render.Widget `starlark:"child,required"`
	XAngle        float64 `starlark:"x_angle"`
	YAngle        float64 `starlark:"y_angle"`
}

func (s Shear) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	cb := s.Widget.PaintBounds(bounds, frameIdx)

	w := float64(cb.Dx())
	h := float64(cb.Dy())

	nw := w
	nh := h

	if s.XAngle != 0 {
		nw += h * math.Abs(math.Tan(gg.Radians(s.XAngle)))
	}

	if s.YAngle != 0 {
		nh += w * math.Abs(math.Tan(gg.Radians(s.YAngle)))
	}

	return image.Rect(0, 0, int(math.Ceil(nw)), int(math.Ceil(nh)))
}

func (s Shear) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	var sx, sy float64
	if s.XAngle != 0 {
		sx = math.Tan(gg.Radians(s.XAngle))
	}
	if s.YAngle != 0 {
		sy = math.Tan(gg.Radians(s.YAngle))
	}

	paintWithTransform(dc, s.Widget, bounds, frameIdx, func(dc *gg.Context) {
		dc.Shear(sx, sy)
	})
}
