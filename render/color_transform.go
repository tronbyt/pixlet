package render

import (
	"image"
	"image/color"
	"math"

	"github.com/tronbyt/gg"
)

// Luminance weights for RGB to grayscale conversion (ITU-R BT.601 standard)
const (
	luminanceRed   = 0.299
	luminanceGreen = 0.587
	luminanceBlue  = 0.114
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
t.Child.Paint(tempDC, image.Rect(0, 0, childBounds.Dx(), childBounds.Dy()), frameIdx)

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
	brightness := normalizeTransformValue(t.Brightness)
	saturation := normalizeTransformValue(t.Saturation)
	opacity := normalizeTransformValue(t.Opacity)
	hueRotate := t.HueRotate

	// Pre-calculate flags for which transformations are active
	applyBrightness := brightness != 1.0
	applySaturation := saturation != 1.0
	applyHueRotate := hueRotate != 0
	applyOpacity := opacity != 1.0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()

			// Skip fully transparent pixels if opacity transformation isn't reducing it further
			if a == 0 && opacity >= 1.0 {
				result.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
				continue
			}

			// Convert to float for calculations
			rf, gf, bf, af := rgba16ToFloat(r, g, b, a)

			// Apply invert
			if t.Invert {
				rf = 1.0 - rf
				gf = 1.0 - gf
				bf = 1.0 - bf
			}

			// Apply brightness
			if applyBrightness {
				rf *= brightness
				gf *= brightness
				bf *= brightness
			}

			// Apply saturation
			if applySaturation {
				// Convert to grayscale using standard luminance weights
				gray := luminanceRed*rf + luminanceGreen*gf + luminanceBlue*bf
				// Interpolate between grayscale and original
				rf = gray + saturation*(rf-gray)
				gf = gray + saturation*(gf-gray)
				bf = gray + saturation*(bf-gray)
			}

			// Apply hue rotation
			if applyHueRotate {
				rf, gf, bf = rotateHue(rf, gf, bf, hueRotate)
			}

			// Apply tint
			if t.Tint != nil {
				tr, tg, tb, _ := t.Tint.RGBA()
				trf, tgf, tbf, _ := rgba16ToFloat(tr, tg, tb, 0)

				// Blend original with tint (multiply blend mode)
				rf *= trf
				gf *= tgf
				bf *= tbf
			}

			// Apply opacity
			if applyOpacity {
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

// rotateHue rotates the hue of an RGB color by the specified degrees.
// This is done by converting to HSL color space, rotating the hue component,
// and converting back to RGB.
func rotateHue(r, g, b, degrees float64) (float64, float64, float64) {
	// Convert RGB to HSL
	h, s, l := rgbToHSL(r, g, b)

	// Rotate hue and normalize to [0, 360)
	h = math.Mod(h+degrees, 360)
	if h < 0 {
		h += 360
	}

	// Convert back to RGB
	return hslToRGB(h, s, l)
}

// rgbToHSL converts RGB color space (each component 0-1) to HSL color space
// (H: 0-360 degrees, S: 0-1, L: 0-1). This uses the standard HSL conversion algorithm.
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

// hslToRGB converts HSL color space (H: 0-360 degrees, S: 0-1, L: 0-1)
// back to RGB color space (each component 0-1). This uses the standard HSL to RGB
// conversion algorithm.
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

// hueToRGB is a helper function for HSL to RGB conversion that calculates
// the RGB component value based on the temporary values p and q, and the
// adjusted hue value t.
func hueToRGB(p, q, t float64) float64 {
	// Normalize t to [0, 360)
	t = math.Mod(t, 360)
	if t < 0 {
		t += 360
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

// normalizeTransformValue returns 1.0 (no change) if the value is negative (indicating "not set"),
// otherwise returns the value as-is. This allows 0 to be a valid transformation value while
// using negative values as a sentinel for "use default".
func normalizeTransformValue(value float64) float64 {
	if value < 0 {
		return 1.0
	}
	return value
}

// rgba16ToFloat converts 16-bit RGBA values (as returned by color.Color.RGBA())
// to normalized float64 values in the range [0, 1] for easier color manipulation.
func rgba16ToFloat(r, g, b, a uint32) (float64, float64, float64, float64) {
	return float64(r>>8) / 255.0, float64(g>>8) / 255.0, float64(b>>8) / 255.0, float64(a>>8) / 255.0
}

// clampFloat clamps a float64 value to the range [0, 1] and converts it to
// a uint8 value in the range [0, 255] for use in color.RGBA.
func clampFloat(f float64) uint8 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 255
	}
	return uint8(f * 255)
}
