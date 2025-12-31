package render

import (
	"image"
	"image/color"
	"testing"
)

func TestColorTransformBrightness(t *testing.T) {
	// Create a simple red box
	box := &Box{
		Width:  10,
		Height: 10,
		Color:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
	}

	// Apply brightness = 0 (should make it black)
	transform := &ColorTransform{
		Child:      box,
		Brightness: 0.0,
	}

	bounds := image.Rect(0, 0, 10, 10)
	img := PaintWidget(transform, bounds, 0)

	// Check that the result is black
	c := img.At(5, 5).(color.RGBA)
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("Expected black (0,0,0), got (%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestColorTransformSaturation(t *testing.T) {
	// Create a simple red box
	box := &Box{
		Width:  10,
		Height: 10,
		Color:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
	}

	// Apply saturation = 0 (should make it grayscale)
	transform := &ColorTransform{
		Child:      box,
		Saturation: 0.0,
	}

	bounds := image.Rect(0, 0, 10, 10)
	img := PaintWidget(transform, bounds, 0)

	// Check that R, G, B are equal (grayscale)
	c := img.At(5, 5).(color.RGBA)
	if c.R != c.G || c.G != c.B {
		t.Errorf("Expected grayscale (R=G=B), got (%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestColorTransformOpacity(t *testing.T) {
	// Create a simple red box
	box := &Box{
		Width:  10,
		Height: 10,
		Color:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
	}

	// Apply opacity = 0.5
	transform := &ColorTransform{
		Child:   box,
		Opacity: 0.5,
	}

	bounds := image.Rect(0, 0, 10, 10)
	img := PaintWidget(transform, bounds, 0)

	// Check that alpha is approximately half
	c := img.At(5, 5).(color.RGBA)
	expectedAlpha := uint8(127) // 255 * 0.5 ≈ 127
	if c.A < expectedAlpha-5 || c.A > expectedAlpha+5 {
		t.Errorf("Expected alpha around %d, got %d", expectedAlpha, c.A)
	}
}

func TestColorTransformInvert(t *testing.T) {
	// Create a simple red box
	box := &Box{
		Width:  10,
		Height: 10,
		Color:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
	}

	// Apply invert
	transform := &ColorTransform{
		Child:      box,
		Brightness: -1, // not set, will default to 1.0
		Saturation: -1, // not set, will default to 1.0
		Opacity:    -1, // not set, will default to 1.0
		Invert:     true,
	}

	bounds := image.Rect(0, 0, 10, 10)
	img := PaintWidget(transform, bounds, 0)

	// Check that red became cyan (inverted red is cyan)
	c := img.At(5, 5).(color.RGBA)
	if c.R != 0 || c.G != 255 || c.B != 255 {
		t.Errorf("Expected cyan (0,255,255), got (%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestColorTransformTint(t *testing.T) {
	// Create a simple white box
	box := &Box{
		Width:  10,
		Height: 10,
		Color:  color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}

	// Apply blue tint
	transform := &ColorTransform{
		Child:      box,
		Brightness: -1, // not set, will default to 1.0
		Saturation: -1, // not set, will default to 1.0
		Opacity:    -1, // not set, will default to 1.0
		Tint:       color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}

	bounds := image.Rect(0, 0, 10, 10)
	img := PaintWidget(transform, bounds, 0)

	// Check that the result is blue-ish (white * blue = blue)
	c := img.At(5, 5).(color.RGBA)
	if c.B == 0 || c.R != 0 || c.G != 0 {
		t.Errorf("Expected blue tint, got (%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestColorTransformNoColorTransformations(t *testing.T) {
	// Create a simple red box
	box := &Box{
		Width:  10,
		Height: 10,
		Color:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
	}

	// ColorTransform with no transformations (should pass through)
	transform := &ColorTransform{
		Child:      box,
		Brightness: -1, // not set, will default to 1.0
		Saturation: -1, // not set, will default to 1.0
		Opacity:    -1, // not set, will default to 1.0
	}

	bounds := image.Rect(0, 0, 10, 10)
	img := PaintWidget(transform, bounds, 0)

	// Check that the result is still red
	c := img.At(5, 5).(color.RGBA)
	if c.R != 255 || c.G != 0 || c.B != 0 {
		t.Errorf("Expected red (255,0,0), got (%d,%d,%d)", c.R, c.G, c.B)
	}
}

func TestColorTransformFrameCount(t *testing.T) {
	// Create a box with a child
	box := &Box{
		Width:  10,
		Height: 10,
		Color:  color.RGBA{R: 255, G: 0, B: 0, A: 255},
	}

	transform := &ColorTransform{
		Child: box,
	}

	bounds := image.Rect(0, 0, 10, 10)
	frameCount := transform.FrameCount(bounds)

	if frameCount != 1 {
		t.Errorf("Expected frame count 1, got %d", frameCount)
	}
}
