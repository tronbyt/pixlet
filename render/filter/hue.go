package filter

import (
	"image"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Hue adjusts the hue of the child widget.
//
// DOC(Widget): The widget to adjust hue for.
// DOC(Change): The amount to change hue by in degrees.
//
// EXAMPLE BEGIN
//
//	filter.Hue(
//	    child = render.Image(src="...", width=64, height=64),
//	    change = 180.0,
//	)
//
// EXAMPLE END.
type Hue struct {
	render.Widget `starlark:"child,required"`

	Change float64 `starlark:"change,required"`
}

func (h Hue) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, h.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		// bild Hue expects int for degrees, but starlark passes float. Casting to int.
		return adjust.Hue(img, int(h.Change))
	})
}
