package filter

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Emboss applies an emboss filter to the child widget.
//
// Example:
//
//	filter.Emboss(
//	    child = render.Image(src="...", width=64, height=64),
//	)
type Emboss struct {
	// The widget to emboss.
	render.Widget `starlark:"child,required"`
}

func (e Emboss) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, e.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return effect.Emboss(img)
	})
}
