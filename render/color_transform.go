package render

import (
	"image"
	"image/color"
	"math"

	"github.com/tronbyt/gg"
)

// ColorTransform applies color transformations to its child widget.
//
// ColorTransform allows you to modify the appearance of any widget by applying
// various color and appearance effects. This is useful for creating effects
// like silhouettes, grayscale images, color overlays, and more.
//
// All transformation parameters are optional and can be combined.
//
// DOC(Child): Widget to apply transformations to
// DOC(Brightness): Brightness multiplier (0=black, 1=normal, >1=brighter)
// DOC(Saturation): Color saturation (0=grayscale, 1=normal)
// DOC(HueRotate): Hue rotation in degrees (0-360)
// DOC(Opacity): Opacity/transparency (0=invisible, 1=opaque)
// DOC(Invert): Whether to invert all colors
// DOC(Tint): Color to blend with the image
//
// EXAMPLE BEGIN
//
//	# Create a black silhouette
//	render.ColorTransform(
//	    child=render.Image(src=icon_data),
//	    brightness=0.0,
//	)
//
// # EXAMPLE END
//
// EXAMPLE BEGIN
//
//	# Grayscale with transparency
//	render.ColorTransform(
//	    child=render.Text("Hello"),
//	    saturation=0.0,
//	    opacity=0.5,
//	)
//
// # EXAMPLE END
//
// EXAMPLE BEGIN
//
//	# Red tint with hue shift
//	render.ColorTransform(
//	    child=render.Box(width=20, height=20, color="#0f0"),
//	    tint="#ff0000",
//	    hue_rotate=180,
//	)
//
// EXAMPLE END
type ColorTransform struct {
	Child      Widget      `starlark:"child,required"`
	Brightness float64     `starlark:"brightness"`
	Saturation float64     `starlark:"saturation"`
	HueRotate  float64     `starlark:"hue_rotate"`
	Opacity    float64     `starlark:"opacity"`
	Invert     bool        `starlark:"invert"`
	Tint       color.Color `starlark:"tint"`
}

func (t *ColorTransform) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	if t.Child == nil {
		return image.Rectangle{}
	}
	return t.Child.PaintBounds(bounds, frameIdx)
}

func (t *ColorTransform) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	if t.Child == nil {
		return
	}

	// Check if any transformations are actually applied
	// Brightness, Saturation, Opacity: < 0 means "not set" (defaults to 1.0 in applyColorTransformations)
	// HueRotate: 0 is "no change"
	// Invert: false is "no change"
	// Tint: nil is "no change"
	needsColorTransform := (t.Brightness >= 0 && t.Brightness != 1.0) ||
		(t.Saturation >= 0 && t.Saturation != 1.0) ||
		t.HueRotate != 0 ||
		(t.Opacity >= 0 && t.Opacity != 1.0) ||
		t.Invert || t.Tint != nil

	if !needsColorTransform {
		// No transformations, just paint the child directly
		t.Child.Paint(dc, bounds, frameIdx)
		return
	}

	// Get child bounds
	childBounds := t.Child.PaintBounds(bounds, frameIdx)
	if childBounds.Dx() <= 0 || childBounds.Dy() <= 0 {
		return
	}

	// Create temporary context to render child
	tempDC := gg.NewContext(childBounds.Dx(), childBounds.Dy())
	t.Child.Paint(tempDC, bounds, frameIdx)

	// Get the rendered image
	img := tempDC.Image()

	// Apply transformations
	transformed := t.applyColorTransformations(img)

	// Draw transformed image to main context
	dc.DrawImage(transformed, 0, 0)
}

func (t *ColorTransform) FrameCount(bounds image.Rectangle) int {
	if t.Child != nil {
		return t.Child.FrameCount(bounds)
	}
	return 1
}

// applyColorTransformations applies all color transformations to the image
func (t *ColorTransform) applyColorTransformations(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Get transformation values with defaults
	// Negative values mean "not set", use 1.0 as default (no change)
	// 0 is a valid value: 0 brightness = black, 0 saturation = grayscale, 0 opacity = transparent
	brightness := t.Brightness
	if brightness < 0 {
		brightness = 1.0
	}
	saturation := t.Saturation
	if saturation < 0 {
		saturation = 1.0
	}
	hueRotate := t.HueRotate
	opacity := t.Opacity
	if opacity < 0 {
		opacity = 1.0
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()

			// Convert from 16-bit to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// Convert to float for calculations
			rf := float64(r8) / 255.0
			gf := float64(g8) / 255.0
			bf := float64(b8) / 255.0
			af := float64(a8) / 255.0

			// Apply invert
			if t.Invert {
				rf = 1.0 - rf
				gf = 1.0 - gf
				bf = 1.0 - bf
			}

			// Apply brightness
			if brightness != 1.0 {
				rf *= brightness
				gf *= brightness
				bf *= brightness
			}

			// Apply saturation
			if saturation != 1.0 {
				// Convert to grayscale using luminance weights
				gray := 0.299*rf + 0.587*gf + 0.114*bf
				// Interpolate between grayscale and original
				rf = gray + saturation*(rf-gray)
				gf = gray + saturation*(gf-gray)
				bf = gray + saturation*(bf-gray)
			}

			// Apply hue rotation
			if hueRotate != 0 {
				rf, gf, bf = rotateHue(rf, gf, bf, hueRotate)
			}

			// Apply tint
			if t.Tint != nil {
				tr, tg, tb, _ := t.Tint.RGBA()
				trf := float64(tr>>8) / 255.0
				tgf := float64(tg>>8) / 255.0
				tbf := float64(tb>>8) / 255.0

				// Blend original with tint (multiply blend mode)
				rf *= trf
				gf *= tgf
				bf *= tbf
			}

			// Apply opacity
			if opacity != 1.0 {
				af *= opacity
			}

			// Clamp and convert back to uint8
			result.SetRGBA(x, y, color.RGBA{
				R: clampFloat(rf),
				G: clampFloat(gf),
				B: clampFloat(bf),
				A: clampFloat(af),
			})
		}
	}

	return result
}

// rotateHue rotates the hue of an RGB color by the given degrees
func rotateHue(r, g, b, degrees float64) (float64, float64, float64) {
	// Convert RGB to HSL
	h, s, l := rgbToHSL(r, g, b)

	// Rotate hue
	h += degrees
	for h >= 360 {
		h -= 360
	}
	for h < 0 {
		h += 360
	}

	// Convert back to RGB
	return hslToRGB(h, s, l)
}

// rgbToHSL converts RGB (0-1) to HSL (H: 0-360, S: 0-1, L: 0-1)
func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	l = (max + min) / 2

	if max == min {
		// Achromatic
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6
			}
		case g:
			h = (b-r)/d + 2
		case b:
			h = (r-g)/d + 4
		}
		h *= 60
	}

	return h, s, l
}

// hslToRGB converts HSL (H: 0-360, S: 0-1, L: 0-1) to RGB (0-1)
func hslToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		// Achromatic
		r = l
		g = l
		b = l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRGB(p, q, h+120)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-120)
	}

	return r, g, b
}

// hueToRGB is a helper function for HSL to RGB conversion
func hueToRGB(p, q, t float64) float64 {
	for t < 0 {
		t += 360
	}
	for t >= 360 {
		t -= 360
	}

	if t < 60 {
		return p + (q-p)*t/60
	}
	if t < 180 {
		return q
	}
	if t < 240 {
		return p + (q-p)*(240-t)/60
	}
	return p
}

// clampFloat clamps a float64 value to [0, 1] and converts to uint8
func clampFloat(f float64) uint8 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 255
	}
	return uint8(f * 255)
}
