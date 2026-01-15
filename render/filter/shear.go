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
//	    y_angle = 0.0,
//	)
//
// EXAMPLE END
type Shear struct {
	render.Widget `starlark:"child,required"`
	XAngle        float64 `starlark:"x_angle,required"`
	YAngle        float64 `starlark:"y_angle,required"`
}

func (s Shear) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	cb := s.Widget.PaintBounds(bounds, frameIdx)

	w := float64(cb.Dx())
	h := float64(cb.Dy())

	nw := w
	nh := h

	if s.XAngle != 0 {
		rad := s.XAngle * math.Pi / 180
		nw += h * math.Abs(math.Tan(rad))
	}

	if s.YAngle != 0 {
		rad := s.YAngle * math.Pi / 180
		nh += w * math.Abs(math.Tan(rad))
	}

	return image.Rect(0, 0, int(math.Ceil(nw)), int(math.Ceil(nh)))
}

func (s Shear) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	cb := s.Widget.PaintBounds(bounds, frameIdx)

	// Calculate center of the provided (expanded) bounds
	cx := float64(bounds.Min.X) + float64(bounds.Dx())/2.0
	cy := float64(bounds.Min.Y) + float64(bounds.Dy())/2.0

	sx := 0.0
	sy := 0.0

	if s.XAngle != 0 {
		sx = -math.Tan(s.XAngle * math.Pi / 180)
	}
	if s.YAngle != 0 {
		sy = -math.Tan(s.YAngle * math.Pi / 180)
	}

	dc.Push()

	// Move to center, shear, then move back by half the child's size.
	dc.Translate(cx, cy)
	dc.Shear(sx, sy)
	dc.Translate(float64(-cb.Dx())/2.0, float64(-cb.Dy())/2.0)

	// Paint child at (0,0) relative to the transformed origin
	s.Widget.Paint(dc, image.Rect(0, 0, cb.Dx(), cb.Dy()), frameIdx)
	dc.Pop()
}
