package filter

import (
	"image"

	"github.com/anthonynsimon/bild/transform"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// FlipHorizontal flips the child widget horizontally.
//
// DOC(Widget): The widget to flip.
//
// EXAMPLE BEGIN
//
//	filter.FlipHorizontal(
//	    child = render.Image(src="...", width=64, height=64),
//	)
//
// EXAMPLE END.
type FlipHorizontal struct {
	render.Widget `starlark:"child,required"`
}

func (f FlipHorizontal) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, f.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return transform.FlipH(img)
	})
}
