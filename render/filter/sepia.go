package filter

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Sepia applies a sepia filter to the child widget.
//
// DOC(Widget): The widget to apply sepia to.
//
// EXAMPLE BEGIN
//
//	filter.Sepia(
//	    child = render.Image(src="...", width=64, height=64),
//	)
//
// EXAMPLE END
type Sepia struct {
	render.Widget `starlark:"child,required"`
}

func (s Sepia) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, s.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return effect.Sepia(img)
	})
}
