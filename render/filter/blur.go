package filter

import (
	"image"
	"math"

	"github.com/anthonynsimon/bild/blur"
	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// Blur applies a Gaussian blur to the child widget.
//
// DOC(Widget): The widget to apply the blur to.
// DOC(Radius): The radius of the Gaussian blur.
//
// EXAMPLE BEGIN
//
//	filter.Blur(
//	    child = render.Image(src="...", width=64, height=64),
//	    radius = 2.0,
//	)
//
// EXAMPLE END.
type Blur struct {
	render.Widget `starlark:"child,required"`

	Radius float64 `starlark:"radius,required"`
}

func (b Blur) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	cb := b.Widget.PaintBounds(bounds, frameIdx)
	padding := int(math.Ceil(b.Radius * 3))
	return image.Rect(0, 0, cb.Dx()+2*padding, cb.Dy()+2*padding)
}

func (b Blur) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	// Get the bounds of the child widget
	cb := b.Widget.PaintBounds(bounds, frameIdx)

	// Add padding to the temporary image to accommodate the blur
	// 3 * radius is a safe bet for Gaussian blur
	padding := int(math.Ceil(b.Radius * 3))
	w := cb.Dx() + 2*padding
	h := cb.Dy() + 2*padding

	tmp := image.NewNRGBA(image.Rect(0, 0, w, h))
	dc2 := gg.NewContextForImage(tmp)

	// Paint the child into the temporary context, offset by padding
	dc2.Push()
	dc2.Translate(float64(padding), float64(padding))
	b.Widget.Paint(dc2, image.Rect(0, 0, cb.Dx(), cb.Dy()), frameIdx)
	dc2.Pop()

	// Apply the blur
	img := blur.Gaussian(dc2.Image(), b.Radius)

	// Draw the result centered in the bounds
	dx := (bounds.Dx() - img.Bounds().Dx()) / 2
	dy := (bounds.Dy() - img.Bounds().Dy()) / 2
	dc.DrawImage(img, dx, dy)
}
