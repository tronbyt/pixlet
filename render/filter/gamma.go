package filter

import (
	"image"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Gamma applies gamma correction to the child widget.
//
// DOC(Widget): The widget to apply gamma correction to.
// DOC(Gamma): The gamma value. 1.0 is no change, < 1.0 darkens, > 1.0 lightens.
//
// EXAMPLE BEGIN
//
//	filter.Gamma(
//	    child = render.Image(src="...", width=64, height=64),
//	    gamma = 0.5,
//	)
//
// EXAMPLE END
type Gamma struct {
	render.Widget `starlark:"child,required"`
	Gamma         float64 `starlark:"gamma,required"`
}

func (g Gamma) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, g.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return adjust.Gamma(img, g.Gamma)
	})
}
