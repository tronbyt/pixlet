package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/internal/colorutil"
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

func TestImageColorPixelPercentage(t *testing.T) {
	strong := color.NRGBA{R: 136, G: 221, B: 238, A: 255}
	faint := color.NRGBA{R: 99, G: 97, B: 89, A: 20}
	im := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	im.SetNRGBA(0, 0, strong) // Outside the visible middle half.
	im.SetNRGBA(2, 2, strong)
	im.SetNRGBA(3, 2, faint)
	var encoded bytes.Buffer
	require.NoError(t, png.Encode(&encoded, im))
	w := &Image{Src: encoded.Bytes(), Width: 2, Height: 2, RetainOriginal: true}
	require.NoError(t, w.Init(nil))
	palette := []color.Color{strong, strong}
	pct, err := w.ColorPixelPercentage(palette, image.Rect(0, 2, 8, 6), true)
	require.NoError(t, err)
	require.Equal(t, 100.0/32, pct)
	pct, err = w.ColorPixelPercentage(palette, image.Rect(-10, 2, 8, 6), true)
	require.NoError(t, err)
	require.Equal(t, 100.0/32, pct, "clip both numerator and denominator")
	pct, err = w.ColorPixelPercentage(palette, image.Rect(20, 20, 30, 30), true)
	require.NoError(t, err)
	require.Zero(t, pct)
	pct, err = w.ColorPixelPercentage(nil, im.Bounds(), true)
	require.NoError(t, err)
	require.Zero(t, pct)

	// Enabling source measurements must not change rendered pixels.
	plain := &Image{Src: encoded.Bytes(), Width: 2, Height: 2}
	require.NoError(t, plain.Init(nil))
	require.Equal(t, plain.imgs, w.imgs)
	require.Nil(t, plain.original)
	_, err = plain.ColorPixelPercentage(palette, im.Bounds(), true)
	require.ErrorContains(t, err, "retain_original=True")
}

func TestImageColorPixelPercentageAlpha(t *testing.T) {
	wanted := color.NRGBA{R: 100, G: 120, B: 140, A: 20}
	im := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	im.SetNRGBA(0, 0, wanted)
	im.SetNRGBA(1, 0, color.NRGBA{R: 100, G: 120, B: 140, A: 255})
	w := &Image{imgs: []image.Image{im}}
	pct, err := w.ColorPixelPercentage([]color.Color{wanted}, im.Bounds(), false)
	require.NoError(t, err)
	require.Equal(t, 50.0, pct, "RGBA matching includes alpha")
}

func TestImageColorPixelPercentageWrappedColor(t *testing.T) {
	wanted := color.NRGBA{R: 1, A: 1}
	im := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	im.SetNRGBA(0, 0, wanted)
	w := &Image{imgs: []image.Image{im}, original: im}
	for _, original := range []bool{false, true} {
		pct, err := w.ColorPixelPercentage([]color.Color{&colorutil.Color{NRGBA: wanted}}, im.Bounds(), original)
		require.NoError(t, err)
		require.Equal(t, 100.0, pct, "wrapped low-alpha channels must stay exact")
	}
}

func TestImageColorPixelPercentageThresholds(t *testing.T) {
	strong := color.NRGBA{R: 136, G: 221, B: 238, A: 255}
	faint := color.NRGBA{R: 222, G: 208, B: 151, A: 190}
	for _, count := range []int{0, 49, 50, 51, 199, 200, 201, 1000} {
		im := image.NewNRGBA(image.Rect(0, 0, 100, 10))
		for i := range 1000 {
			c := faint // Even widespread faint returns must not count.
			if i < count {
				c = strong
			}
			im.SetNRGBA(i%100, i/100, c)
		}
		w := &Image{imgs: []image.Image{im}, original: im}
		pct, err := w.ColorPixelPercentage([]color.Color{strong}, im.Bounds(), true)
		require.NoError(t, err)
		require.InDelta(t, float64(count)/10, pct, 1e-10)
		require.Equal(t, count >= 50, pct >= 5, "5%% threshold, count=%d", count)
		require.Equal(t, count >= 200, pct >= 20, "20%% threshold, count=%d", count)
	}
}
