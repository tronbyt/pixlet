package filter

import (
	"image"
	"math"

	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Rotate rotates the child widget by the specified angle.
//
// DOC(Widget): The widget to rotate.
// DOC(Angle): The angle to rotate by in degrees.
//
// EXAMPLE BEGIN
//
//	filter.Rotate(
//	    child = render.Image(src="...", width=64, height=64),
//	    angle = 10.0,
//	)
//
// EXAMPLE END.
type Rotate struct {
	render.Widget `starlark:"child,required"`

	Angle float64 `starlark:"angle,required"`
}

func (r Rotate) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	cb := r.Widget.PaintBounds(bounds, frameIdx)

	// Calculate rotated bounds
	// Width' = Width * |cos A| + Height * |sin A|
	// Height' = Width * |sin A| + Height * |cos A|
	rad := gg.Radians(r.Angle)
	cos := math.Abs(math.Cos(rad))
	sin := math.Abs(math.Sin(rad))

	w := float64(cb.Dx())
	h := float64(cb.Dy())

	nw := int(math.Ceil(w*cos + h*sin))
	nh := int(math.Ceil(w*sin + h*cos))

	return image.Rect(0, 0, nw, nh)
}

func (r Rotate) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paintWithTransform(dc, r.Widget, bounds, frameIdx, func(dc *gg.Context) {
		dc.Rotate(gg.Radians(r.Angle))
	})
}
