package filter

import (
	"image"

	"github.com/anthonynsimon/bild/segment"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Threshold applies a threshold filter to the child widget, making it black and white.
//
// EXAMPLE BEGIN
//
//	filter.Threshold(
//	    child = render.Image(src="...", width=64, height=64),
//	    level = 128.0,
//	)
//
// EXAMPLE END.
type Threshold struct {
	// The widget to apply threshold to.
	render.Widget `starlark:"child,required"`

	// The threshold level, from 0 to 255.
	Level float64 `starlark:"level,required"`
}

func (t Threshold) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	paint(dc, t.Widget, bounds, frameIdx, func(img image.Image) image.Image {
		// bild Threshold expects uint8
		return segment.Threshold(img, uint8(t.Level))
	})
}
