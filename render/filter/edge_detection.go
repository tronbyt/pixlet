package filter

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// EdgeDetection applies an edge detection filter to the child widget.
//
// DOC(Widget): The widget to detect edges on.
// DOC(Radius): The radius of the edge detection kernel.
//
// EXAMPLE BEGIN
//
//	filter.EdgeDetection(
//	    child = render.Image(src="...", width=64, height=64),
//	    radius = 2.0,
//	)
//
// EXAMPLE END.
type EdgeDetection struct {
	render.Widget `starlark:"child,required"`

	Radius float64 `starlark:"radius,required"`
}

func (e EdgeDetection) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, e.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return effect.EdgeDetection(img, e.Radius)
	})
}
