package filter

import (
	"image"

	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render"
)

// paint is a helper function to apply a filter to a child widget.
// It creates a temporary image of the child's size, paints the child,
// applies the filter function, and then draws the result onto the
// destination context.
//
// If the filter function returns an image with different dimensions
// than the child, the result is drawn centered relative to the
// child's original position.
func paint(dc *gg.Context, w render.Widget, bounds image.Rectangle, frameIdx int, fn func(image.Image) image.Image) {
	// Get the bounds of the child widget
	cb := w.PaintBounds(bounds, frameIdx)

	// Create a temporary context of the exact size needed for the child
	tmp := image.NewNRGBA(image.Rect(0, 0, cb.Dx(), cb.Dy()))
	dc2 := gg.NewContextForImage(tmp)

	// Paint the child into the temporary context
	// We use local coordinates (0, 0) for the child since dc2 is tight
	w.Paint(dc2, image.Rect(0, 0, cb.Dx(), cb.Dy()), frameIdx)

	// Apply the filter function
	res := fn(dc2.Image())

	// Calculate the position to draw the result to keep it centered
	// relative to the bounds.
	dx := (bounds.Dx() - res.Bounds().Dx()) / 2
	dy := (bounds.Dy() - res.Bounds().Dy()) / 2

	// Draw the result onto the main context
	dc.DrawImage(res, dx, dy)
}
