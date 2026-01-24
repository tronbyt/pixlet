package filter

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Grayscale converts the child widget to grayscale.
//
// DOC(Widget): The widget to convert to grayscale.
//
// EXAMPLE BEGIN
//
//	filter.Grayscale(
//	    child = render.Image(src="...", width=64, height=64),
//	)
//
// EXAMPLE END.
type Grayscale struct {
	render.Widget `starlark:"child,required"`
}

func (g Grayscale) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, g.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return effect.Grayscale(img)
	})
}
