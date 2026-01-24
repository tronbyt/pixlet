package filter

import (
	"image"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Saturation adjusts the saturation of the child widget.
//
// DOC(Widget): The widget to adjust saturation for.
// DOC(Factor): The factor to adjust saturation by. 0.0 is grayscale, 1.0 is no change, > 1.0 increases saturation.
//
// EXAMPLE BEGIN
//
//	filter.Saturation(
//	    child = render.Image(src="...", width=64, height=64),
//	    factor = 1,
//	)
//
// EXAMPLE END.
type Saturation struct {
	render.Widget `starlark:"child,required"`

	Factor float64 `starlark:"factor,required"`
}

func (s Saturation) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, s.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return adjust.Saturation(img, s.Factor)
	})
}
