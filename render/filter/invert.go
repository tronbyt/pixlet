package filter

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Invert inverts the colors of the child widget.
//
// DOC(Widget): The widget to invert.
//
// EXAMPLE BEGIN
//
//	filter.Invert(
//	    child = render.Image(src="...", width=64, height=64),
//	)
//
// EXAMPLE END.
type Invert struct {
	render.Widget `starlark:"child,required"`
}

func (i Invert) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, i.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return effect.Invert(img)
	})
}
