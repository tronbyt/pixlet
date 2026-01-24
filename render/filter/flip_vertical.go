package filter

import (
	"image"

	"github.com/anthonynsimon/bild/transform"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// FlipVertical flips the child widget vertically.
//
// DOC(Widget): The widget to flip.
//
// EXAMPLE BEGIN
//
//	filter.FlipVertical(
//	    child = render.Image(src="...", width=64, height=64),
//	)
//
// EXAMPLE END.
type FlipVertical struct {
	render.Widget `starlark:"child,required"`
}

func (f FlipVertical) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, f.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return transform.FlipV(img)
	})
}
