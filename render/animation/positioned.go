package animation

import (
	"image"
	"math"

	"github.com/tronbyt/gg"

	"github.com/tronbyt/pixlet/render"
)

// AnimatedPositioned animates a widget from start to end coordinates.
//
// **DEPRECATED**: Please use `animation.Transformation` instead.
type AnimatedPositioned struct {
	// Widget to animate
	Child render.Widget `starlark:"child,required"`
	// Horizontal start coordinate
	XStart int `starlark:"x_start"`
	// Horizontal end coordinate
	XEnd int `starlark:"x_end"`
	// Vertical start coordinate
	YStart int `starlark:"y_start"`
	// Vertical end coordinate
	YEnd int `starlark:"y_end"`
	// Duration of animation in frames
	Duration int `starlark:"duration,required"`
	// Easing curve to use, default is 'linear'
	Curve Curve `starlark:"curve,required"`
	// Delay before animation in frames
	Delay int `starlark:"delay"`
	// Delay after animation in frames.
	Hold int `starlark:"hold"`
}

func (o AnimatedPositioned) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	return bounds
}

func (o AnimatedPositioned) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	var position float64

	if frameIdx < o.Delay {
		position = 0.0
	} else if frameIdx >= o.Delay+o.Duration {
		position = 0.9999999999
	} else {
		position = o.Curve.Transform(float64(frameIdx-o.Delay) / float64(o.Duration))
	}

	dx := 1
	if o.XStart > o.XEnd {
		dx = -1
	}
	dy := 1
	if o.YStart > o.YEnd {
		dy = -1
	}

	sx := int(math.Ceil(math.Abs(float64(o.XEnd-o.XStart)) * position))
	sy := int(math.Ceil(math.Abs(float64(o.YEnd-o.YStart)) * position))

	x := o.XStart + dx*sx
	y := o.YStart + dy*sy

	dc.Push()
	dc.Translate(float64(x), float64(y))
	o.Child.Paint(dc, bounds, frameIdx)
	dc.Pop()
}

func (o AnimatedPositioned) FrameCount(bounds image.Rectangle) int {
	return o.Duration + o.Delay + o.Hold
}
