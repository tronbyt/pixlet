package encode

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"
)

type RenderFilters struct {
	Magnify     int             `json:"magnify,omitempty"`
	ColorFilter ColorFilterType `json:"color_filter,omitempty"`
}

type ColorFilterType string

const (
	ColorNone      ColorFilterType = "none"      // No transformation
	ColorDimmed    ColorFilterType = "dimmed"    // Darkens image uniformly while preserving hue
	ColorRedShift  ColorFilterType = "redshift"  // CCT-derived chromatic adaptation matrix (~3400K target)
	ColorWarm      ColorFilterType = "warm"      // Adds a subtle warm, orange/yellow hue
	ColorSunset    ColorFilterType = "sunset"    // Emulates deep pink/orange of a setting sun
	ColorSepia     ColorFilterType = "sepia"     // Adds a warm, antique brown tone mimicking aged photographs
	ColorVintage   ColorFilterType = "vintage"   // Muted, brown/green nostalgic tones
	ColorDusk      ColorFilterType = "dusk"      // Fades brightness, adds reddish cast
	ColorCool      ColorFilterType = "cool"      // Adds a cool, blue tint
	ColorBW        ColorFilterType = "bw"        // Converts image to perceptual grayscale using luminance weights
	ColorIce       ColorFilterType = "ice"       // Pale desaturation with bluish cast
	ColorMoonlight ColorFilterType = "moonlight" // Dim blue-gray, night lighting effect
	ColorNeon      ColorFilterType = "neon"      // Boosts contrast, magenta-blue cyberpunk
	ColorPastel    ColorFilterType = "pastel"    // Softens tones, gentle highlight boost
)

var colorFilterMatrices = map[ColorFilterType][3][3]float32{
	// === Neutral ===
	// No-op identity matrix, avoids altering original colours
	ColorNone: {
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	},

	// Uniform intensity reduction, (same hue)
	ColorDimmed: {
		{0.25, 0, 0},
		{0, 0.25, 0},
		{0, 0, 0.25},
	},

	// === Warm & Tinted ===
	ColorRedShift: {
		// Chromatic adaptation matrix approximating a D65 → 3400K whitepoint shift.
		// Computed in XYZ space using the Bradford method and projected into linear sRGB.
		// Target whitepoint derived from McCamy's CCT approximation.
		{1.2066, 0.3380, 0.0383},
		{-0.0164, 0.8985, 0.0098},
		{-0.0156, -0.0500, 0.4201},
	},

	// Slight red/yellow push, feels warmer and more "golden hour"
	ColorWarm: {
		{1.1, 0.05, 0.0},
		{0.0, 1.0, 0.0},
		{0.05, 0.0, 0.9},
	},

	// Strong warm pink/red sunset vibe
	ColorSunset: {
		{1.2, 0.2, 0.0},
		{0.1, 1.0, 0.1},
		{0.0, 0.1, 0.6},
	},

	// Warm, antique brownish tones mimicking aged photographs
	ColorSepia: {
		{0.393, 0.769, 0.189},
		{0.349, 0.686, 0.168},
		{0.272, 0.534, 0.131},
	},

	// Classic retro TV look, muted greens and browns
	ColorVintage: {
		{1.0, 0.6, 0.2},
		{0.3, 0.9, 0.2},
		{0.2, 0.4, 0.6},
	},

	// Darkens greens and blues, slight red push for evening tones
	ColorDusk: {
		{1.1, 0.0, 0.2},
		{0.0, 0.8, 0.1},
		{0.0, 0.1, 0.6},
	},

	// === Cool & Desaturated ===
	// Adds subtle blue tint for cooler "tech" or night feel
	ColorCool: {
		{0.9, 0.0, 0.2},
		{0.0, 1.0, 0.0},
		{-0.1, 0.0, 1.1},
	},

	// Perceptual grayscale using luminance weights
	ColorBW: {
		{0.3, 0.59, 0.11},
		{0.3, 0.59, 0.11},
		{0.3, 0.59, 0.11},
	},
	// Desaturated blues
	ColorIce: {
		{0.8, 0.9, 1.0},
		{0.8, 0.9, 1.0},
		{1.0, 1.0, 1.2},
	},

	// Moonlight/night vision vibes
	ColorMoonlight: {
		{0.6, 0.2, 0.4},
		{0.2, 0.7, 0.2},
		{0.3, 0.3, 0.9},
	},

	// === Stylized  ===
	// High contrast, exaggerated hues, magenta-blue cyberpunk/glow aesthetic
	ColorNeon: {
		{0.9, 0.0, 1.1},
		{0.0, 1.0, 0.6},
		{0.2, 0.5, 1.3},
	},

	// Boosts all channels lightly
	ColorPastel: {
		{1.2, 0.1, 0.1},
		{0.1, 1.2, 0.1},
		{0.1, 0.1, 1.2},
	},
}

func (f ColorFilterType) String() string {
	return string(f)
}

func (f *ColorFilterType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*f = ColorFilterType(s)
	return nil
}

func ValidateColorFilter(input ColorFilterType) (ColorFilterType, error) {
	if input == "" {
		return ColorNone, nil
	}
	filterType := ColorFilterType(input)
	if !filterType.IsValid() {
		return "", fmt.Errorf("invalid color filter: %q\nSupported filters: %s",
			input,
			strings.Join(SupportedColorFilters(), ", "),
		)
	}
	return filterType, nil
}

func FromFilterType(f ColorFilterType) (ImageFilter, error) {
	if f == ColorNone {
		return nil, nil // explicit noop skip
	}
	matrix, ok := colorFilterMatrices[f]
	if !ok {
		return nil, fmt.Errorf("unknown color filter: %q", f)
	}
	// log.Printf("Applying filter: %s", f)
	return ColorMatrix(matrix), nil
}

func (f ColorFilterType) IsValid() bool {
	_, ok := colorFilterMatrices[f]
	return ok
}

func SupportedColorFilters() []string {
	keys := make([]string, 0, len(colorFilterMatrices))
	for k := range colorFilterMatrices {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	return keys
}

func (f RenderFilters) String() string {
	return fmt.Sprintf("Magnify=%d, ColorFilter=%q", f.Magnify, f.ColorFilter)
}

// Applies a sequence of ImageFilters in order.
func Chain(filters ...ImageFilter) ImageFilter {
	return func(img image.Image) (image.Image, error) {
		var err error
		for _, f := range filters {
			img, err = f(img)
			if err != nil {
				return nil, fmt.Errorf("filter failed: %w", err)
			}
		}
		return img, nil
	}
}

// Enlarges an image by an integer factor.
func Magnify(factor int) ImageFilter {
	return func(input image.Image) (image.Image, error) {
		if factor <= 1 {
			return input, nil
		}
		in, ok := input.(*image.RGBA)
		if !ok {
			bounds := input.Bounds()
			tmp := image.NewRGBA(bounds)
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					tmp.Set(x, y, input.At(x, y))
				}
			}
			in = tmp
		}

		out := image.NewRGBA(
			image.Rect(
				0, 0,
				in.Bounds().Dx()*factor,
				in.Bounds().Dy()*factor,
			),
		)

		for x := 0; x < in.Bounds().Dx(); x++ {
			for y := 0; y < in.Bounds().Dy(); y++ {
				px := in.RGBAAt(x, y)
				for dx := 0; dx < factor; dx++ {
					for dy := 0; dy < factor; dy++ {
						out.SetRGBA(x*factor+dx, y*factor+dy, px)
					}
				}
			}
		}
		return out, nil
	}
}

// Apply a 3x3 color transformation matrix to the RGB values of an image.
func ColorMatrix(matrix [3][3]float32) ImageFilter {
	return func(img image.Image) (image.Image, error) {
		bounds := img.Bounds()
		out := image.NewRGBA(bounds)

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r0, g0, b0, a0 := img.At(x, y).RGBA()
				r := float32(r0 >> 8)
				g := float32(g0 >> 8)
				b := float32(b0 >> 8)

				nr := matrix[0][0]*r + matrix[0][1]*g + matrix[0][2]*b
				ng := matrix[1][0]*r + matrix[1][1]*g + matrix[1][2]*b
				nb := matrix[2][0]*r + matrix[2][1]*g + matrix[2][2]*b

				out.Set(x, y, color.RGBA{
					R: clamp(nr),
					G: clamp(ng),
					B: clamp(nb),
					A: uint8(a0 >> 8),
				})
			}
		}
		return out, nil
	}
}

func clamp(f float32) uint8 {
	if f < 0 {
		return 0
	}
	if f > 255 {
		return 255
	}
	return uint8(f)
}
