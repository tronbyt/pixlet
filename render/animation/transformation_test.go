package animation

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/render"
)

func TestTransformationTranslate(t *testing.T) {
	o := Transformation{
		Child: render.Column{
			Children: []render.Widget{
				render.Box{Width: 3, Height: 1, Color: color.RGBA{0xff, 0, 0, 0xff}},
				render.Box{Width: 3, Height: 1, Color: color.RGBA{0, 0xff, 0, 0xff}},
				render.Box{Width: 3, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
			},
		},
		Keyframes: []Keyframe{
			{
				Percentage: Percentage{0.0},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Translate{X: 0.0, Y: 0.0},
				},
			},
			{
				Percentage: Percentage{1.0},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Translate{X: 5.0, Y: 5.0},
				},
			},
		},
		Duration:     6,
		Delay:        0,
		Width:        5,
		Height:       5,
		Origin:       Origin{X: Percentage{0.5}, Y: Percentage{0.5}},
		Direction:    DefaultDirection,
		FillMode:     DefaultFillMode,
		Rounding:     DefaultRounding,
		WaitForChild: false,
	}

	// These frames should show the box moving diagonally out of frame.
	assert.Equal(t, 6, o.FrameCount(image.Rect(0, 0, 5, 5)))

	im := render.PaintWidget(&o, image.Rect(0, 0, 5, 5), 0)
	require.NoError(t, render.CheckImage([]string{
		"rrr..",
		"ggg..",
		"bbb..",
		".....",
		".....",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 5, 5), 1)
	require.NoError(t, render.CheckImage([]string{
		".....",
		".rrr.",
		".ggg.",
		".bbb.",
		".....",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 5, 5), 2)
	require.NoError(t, render.CheckImage([]string{
		".....",
		".....",
		"..rrr",
		"..ggg",
		"..bbb",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 5, 5), 3)
	require.NoError(t, render.CheckImage([]string{
		".....",
		".....",
		".....",
		"...rr",
		"...gg",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 5, 5), 4)
	require.NoError(t, render.CheckImage([]string{
		".....",
		".....",
		".....",
		".....",
		"....r",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 5, 5), 5)
	require.NoError(t, render.CheckImage([]string{
		".....",
		".....",
		".....",
		".....",
		".....",
	}, im))
}

func TestTransformationScale(t *testing.T) {
	o := Transformation{
		// Choosing only red, as scaling will interpolate between colors...
		Child: render.Box{Width: 3, Height: 3, Color: color.RGBA{0xff, 0, 0, 0xff}},
		Keyframes: []Keyframe{
			{
				Percentage: Percentage{0.0},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Scale{X: 1.0, Y: 1.0},
				},
			},
			{
				Percentage: Percentage{0.5},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Scale{X: 2.0, Y: 2.0},
				},
			},
			{
				Percentage: Percentage{1.0},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Scale{X: 3.0, Y: 3.0},
				},
			},
		},
		Duration:     3,
		Delay:        0,
		Width:        9,
		Height:       9,
		Origin:       Origin{X: Percentage{0.0}, Y: Percentage{0.0}},
		Direction:    DefaultDirection,
		FillMode:     DefaultFillMode,
		Rounding:     DefaultRounding,
		WaitForChild: false,
	}

	// These frames should show the box scaling from 1x to 3x.
	assert.Equal(t, 3, o.FrameCount(image.Rect(0, 0, 9, 9)))

	im := render.PaintWidget(&o, image.Rect(0, 0, 9, 9), 0)
	require.NoError(t, render.CheckImage([]string{
		"rrr......",
		"rrr......",
		"rrr......",
		".........",
		".........",
		".........",
		".........",
		".........",
		".........",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 9, 9), 1)
	require.NoError(t, render.CheckImage([]string{
		"rrrrrr...",
		"rrrrrr...",
		"rrrrrr...",
		"rrrrrr...",
		"rrrrrr...",
		"rrrrrr...",
		".........",
		".........",
		".........",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 9, 9), 2)
	require.NoError(t, render.CheckImage([]string{
		"rrrrrrrrr",
		"rrrrrrrrr",
		"rrrrrrrrr",
		"rrrrrrrrr",
		"rrrrrrrrr",
		"rrrrrrrrr",
		"rrrrrrrrr",
		"rrrrrrrrr",
		"rrrrrrrrr",
	}, im))
}

func TestTransformationRotate(t *testing.T) {
	o := Transformation{
		Child: render.Column{
			Children: []render.Widget{
				render.Box{Width: 3, Height: 1, Color: color.RGBA{0xff, 0, 0, 0xff}},
				render.Box{Width: 3, Height: 1, Color: color.RGBA{0, 0xff, 0, 0xff}},
				render.Box{Width: 3, Height: 1, Color: color.RGBA{0, 0, 0xff, 0xff}},
			},
		},
		Keyframes: []Keyframe{
			{
				Percentage: Percentage{0.0},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Rotate{Angle: 0.0},
				},
			},
			{
				Percentage: Percentage{1.0},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Rotate{Angle: 360.0},
				},
			},
		},
		Duration:     5,
		Delay:        0,
		Width:        3,
		Height:       3,
		Origin:       DefaultOrigin,
		Direction:    DefaultDirection,
		FillMode:     DefaultFillMode,
		Rounding:     DefaultRounding,
		WaitForChild: false,
	}

	// These frames should show the box rotating 90 degrees each frame.
	assert.Equal(t, 5, o.FrameCount(image.Rect(0, 0, 3, 3)))

	im := render.PaintWidget(&o, image.Rect(0, 0, 3, 3), 0)
	require.NoError(t, render.CheckImage([]string{
		"rrr",
		"ggg",
		"bbb",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 3, 3), 1)
	require.NoError(t, render.CheckImage([]string{
		"bgr",
		"bgr",
		"bgr",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 3, 3), 2)
	require.NoError(t, render.CheckImage([]string{
		"bbb",
		"ggg",
		"rrr",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 3, 3), 3)
	require.NoError(t, render.CheckImage([]string{
		"rgb",
		"rgb",
		"rgb",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 3, 3), 4)
	require.NoError(t, render.CheckImage([]string{
		"rrr",
		"ggg",
		"bbb",
	}, im))
}

func TestTransformationAll(t *testing.T) {
	// Checking colors with anti-aliasing (when scaling) is more complex.
	ic := render.ImageChecker{Palette: map[string]color.RGBA{
		"█": {0xff, 0xff, 0xff, 0xff},
		"▓": {0x40, 0x40, 0x40, 0x40},
		"▒": {0x20, 0x20, 0x20, 0x20},
		"░": {0x08, 0x08, 0x08, 0x08},
		"⎕": {0x04, 0x04, 0x04, 0x04},
		".": {0, 0, 0, 0},
		"●": {0, 0, 0, 0xff},
		"◉": {0, 0, 0, 0x40},
		"◎": {0, 0, 0, 0x20},
		"○": {0, 0, 0, 0x08},
		"⁘": {0, 0, 0, 0x04},
	}}

	o := Transformation{
		// █.●
		// ...
		// ●.█
		Child: render.Box{
			Child: render.Column{
				Children: []render.Widget{
					render.Row{
						Children: []render.Widget{
							render.Box{Width: 1, Height: 1, Color: color.RGBA{0xff, 0xff, 0xff, 0xff}},
							render.Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0, 0}},
							render.Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0, 0xff}},
						},
					},
					render.Box{Width: 3, Height: 1, Color: color.RGBA{0, 0, 0, 0}},
					render.Row{
						Children: []render.Widget{
							render.Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0, 0xff}},
							render.Box{Width: 1, Height: 1, Color: color.RGBA{0, 0, 0, 0}},
							render.Box{Width: 1, Height: 1, Color: color.RGBA{0xff, 0xff, 0xff, 0xff}},
						},
					},
				},
			}},
		Keyframes: []Keyframe{
			{
				Percentage: Percentage{0.0},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Translate{X: -3.0, Y: -3.0},
					Scale{X: 1.0, Y: 1.0},
					Rotate{Angle: 0.0},
				},
			},
			{
				Percentage: Percentage{0.75},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Translate{X: 0.0, Y: 0.0},
					Scale{X: 1.0, Y: 1.0},
					Rotate{Angle: 270.0},
				},
			},
			{
				Percentage: Percentage{1.0},
				Curve:      LinearCurve{},
				Transforms: []Transform{
					Translate{X: 1.0, Y: 1.0},
					Scale{X: 2.0, Y: 2.0},
					Rotate{Angle: 360.0},
				},
			},
		},
		Duration:     5,
		Delay:        0,
		Width:        9,
		Height:       9,
		Origin:       DefaultOrigin,
		Direction:    DefaultDirection,
		FillMode:     DefaultFillMode,
		Rounding:     DefaultRounding,
		WaitForChild: false,
	}

	// These frames should show the four "corners" being,
	// translated, rotated and in the end scaled to 2x.
	assert.Equal(t, 5, o.FrameCount(image.Rect(0, 0, 9, 9)))

	im := render.PaintWidget(&o, image.Rect(0, 0, 9, 9), 0)
	require.NoError(t, ic.Check([]string{
		"█.●......",
		".........",
		"●.█......",
		".........",
		".........",
		".........",
		".........",
		".........",
		".........",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 9, 9), 1)
	require.NoError(t, ic.Check([]string{
		".........",
		".●.█.....",
		".........",
		".█.●.....",
		".........",
		".........",
		".........",
		".........",
		".........",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 3, 3), 2)
	require.NoError(t, ic.Check([]string{
		".........",
		".........",
		"..█.●....",
		".........",
		"..●.█....",
		".........",
		".........",
		".........",
		".........",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 3, 3), 3)
	require.NoError(t, ic.Check([]string{
		".........",
		".........",
		".........",
		"...●.█...",
		".........",
		"...█.●...",
		".........",
		".........",
		".........",
	}, im))

	im = render.PaintWidget(&o, image.Rect(0, 0, 3, 3), 4)

	require.NoError(t, ic.Check([]string{
		".........",
		".........",
		"..⎕▒░.○◎⁘",
		"..▒█▓.◉●◎",
		"..⎕▒░.○◎⁘",
		".........",
		"..⁘◎○.░▒⎕",
		"..◎●◉.▓█▒",
		"..⁘◎○.░▒⎕",
	}, im))
}

// Regression test: a zero scale factor produced a singular transformation
// matrix, which made x/image/draw panic with "makeslice: len out of range"
// when the child painted an image or text through it.
func TestTransformationDegenerateScale(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 24, 32))
	for i := range img.Pix {
		img.Pix[i] = 0xff
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))

	for _, tc := range []struct {
		name  string
		scale Scale
	}{
		{"x zero", Scale{X: 0.0, Y: 1.0}},
		{"y zero", Scale{X: 1.0, Y: 0.0}},
		{"both zero", Scale{X: 0.0, Y: 0.0}},
		{"x tiny", Scale{X: 1e-12, Y: 1.0}},
		{"nan", Scale{X: math.NaN(), Y: 1.0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			child := &render.Image{Src: buf.Bytes()}
			require.NoError(t, child.Init(nil))

			o := Transformation{
				Child: child,
				Keyframes: []Keyframe{
					{
						Percentage: Percentage{0.0},
						Curve:      LinearCurve{},
						Transforms: []Transform{Translate{X: 0.0, Y: -32.0}, tc.scale},
					},
					{
						Percentage: Percentage{1.0},
						Curve:      LinearCurve{},
						Transforms: []Transform{Translate{X: 0.0, Y: -32.0}, tc.scale},
					},
				},
				Duration:  2,
				Width:     64,
				Height:    32,
				Origin:    Origin{X: Percentage{0.5}, Y: Percentage{0.5}},
				Direction: DefaultDirection,
				FillMode:  DefaultFillMode,
				Rounding:  DefaultRounding,
			}
			require.NoError(t, o.Init(nil))

			require.NotPanics(t, func() {
				render.PaintWidget(&o, image.Rect(0, 0, 64, 32), 0)
			})
		})
	}
}

func TestTransformationSmallScaleStillPaints(t *testing.T) {
	child := render.Box{Width: 4, Height: 4, Color: color.RGBA{0xff, 0, 0, 0xff}}

	o := Transformation{
		Child: child,
		Keyframes: []Keyframe{
			{
				Percentage: Percentage{0.0},
				Curve:      LinearCurve{},
				Transforms: []Transform{Scale{X: 0.5, Y: 1.0}},
			},
		},
		Duration:  1,
		Width:     4,
		Height:    4,
		Origin:    Origin{X: Percentage{0.5}, Y: Percentage{0.5}},
		Direction: DefaultDirection,
		FillMode:  DefaultFillMode,
		Rounding:  DefaultRounding,
	}
	require.NoError(t, o.Init(nil))

	im := render.PaintWidget(&o, image.Rect(0, 0, 4, 4), 0)
	assert.Equal(t, color.RGBA{0xff, 0, 0, 0xff}, im.At(2, 2), "half-scaled child should still be painted")
}
