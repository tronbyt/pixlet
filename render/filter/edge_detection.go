package filter

import (
	"image"

	"github.com/anthonynsimon/bild/effect"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// EdgeDetection applies an edge detection filter to the child widget.
//
// Example:
//
//	filter.EdgeDetection(
//	    child = render.Image(src="...", width=64, height=64),
//	    radius = 2.0,
//	)
type EdgeDetection struct {
	// The widget to detect edges on.
	render.Widget `starlark:"child,required"`

	// The radius of the edge detection kernel.
	Radius float64 `starlark:"radius,required"`
}

func (e EdgeDetection) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, e.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		return effect.EdgeDetection(img, e.Radius)
	})
}
