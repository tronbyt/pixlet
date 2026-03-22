package filter

import (
	"image"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Brightness adjusts the brightness of the child widget.
//
// EXAMPLE BEGIN
//
//	filter.Brightness(
//	    child = render.Image(src="...", width=64, height=64),
//	    change = -0.5,
//	)
//
// EXAMPLE END.
type Brightness struct {
	// The widget to adjust brightness for.
	render.Widget `starlark:"child,required"`

	// The amount to change brightness by. -1.0 is black, 1.0 is white, 0.0 is no change.
	Change float64 `starlark:"change,required"`
}

func (b Brightness) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, b.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return adjust.Brightness(img, b.Change)
	})
}
