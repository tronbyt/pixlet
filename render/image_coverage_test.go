package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageOpaquePixelPercentage(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	im.Set(0, 0, color.NRGBA{A: 255})
	im.Set(1, 0, color.NRGBA{A: 1})
	im.Set(2, 0, color.NRGBA{A: 255})

	widget := &Image{imgs: []image.Image{im}}

	require.Equal(t, 37.5, widget.OpaquePixelPercentage(im.Bounds()))
	require.Equal(t, 50.0, widget.OpaquePixelPercentage(image.Rect(0, 0, 2, 2)))
	require.Equal(t, 50.0, widget.OpaquePixelPercentage(image.Rect(-1, 0, 3, 2)))
	require.Equal(t, 0.0, widget.OpaquePixelPercentage(image.Rect(10, 10, 12, 12)))
}
