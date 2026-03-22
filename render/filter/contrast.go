package filter

import (
	"image"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Contrast adjusts the contrast of the child widget.
//
// EXAMPLE BEGIN
//
//	filter.Contrast(
//	    child = render.Image(src="...", width=64, height=64),
//	    factor = 2.0,
//	)
//
// EXAMPLE END.
type Contrast struct {
	// The widget to adjust contrast for.
	render.Widget `starlark:"child,required"`

	// The factor to adjust contrast by. -1.0 is gray, 1.0 is no change, > 1.0 increases contrast.
	Factor float64 `starlark:"factor,required"`
}

func (c Contrast) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, c.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return adjust.Contrast(img, c.Factor)
	})
}
