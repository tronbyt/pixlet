package filter

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Sharpen sharpens the child widget.
//
// DOC(Widget): The widget to sharpen.
//
// EXAMPLE BEGIN
//
//	filter.Sharpen(
//	    child = render.Image(src="...", width=64, height=64),
//	)
//
// EXAMPLE END.
type Sharpen struct {
	render.Widget `starlark:"child,required"`
}

func (s Sharpen) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, s.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return effect.Sharpen(img)
	})
}
