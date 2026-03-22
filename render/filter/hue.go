package filter

import (
	"image"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Hue adjusts the hue of the child widget.
//
// Example:
//
//	filter.Hue(
//	    child = render.Image(src="...", width=64, height=64),
//	    change = 180.0,
//	)
type Hue struct {
	// The widget to adjust hue for.
	render.Widget `starlark:"child,required"`

	// The amount to change hue by in degrees.
	Change float64 `starlark:"change,required"`
}

func (h Hue) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, h.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		// bild Hue expects int for degrees, but starlark passes float. Casting to int.
		return adjust.Hue(img, int(h.Change))
	})
}
