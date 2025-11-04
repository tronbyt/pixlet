package encode

import (
	"fmt"
	"image"
	"image/color"
)

//go:generate go tool enumer -type ColorFilter -trimprefix Color -transform lower -output filter_string.go

type ColorFilter uint8

const (
	ColorNone ColorFilter = iota
	ColorDimmed
	ColorRedShift
	ColorWarm
	ColorSunset
	ColorSepia
	ColorVintage
	ColorDusk
	ColorCool
	ColorBW
	ColorIce
	ColorMoonlight
	ColorNeon
	ColorPastel
)

var colorFilters = map[ColorFilter][3][3]float32{
	ColorNone: {
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
	},
	ColorDimmed: {
		{0.25, 0, 0},
		{0, 0.25, 0},
		{0, 0, 0.25},
	},
	ColorRedShift: {
		// Chromatic adaptation matrix approximating a D65 → 3400K whitepoint shift.
		// Computed in XYZ space using the Bradford method and projected into linear sRGB.
		// Target whitepoint derived from McCamy's CCT approximation.
		{1.2066, 0.3380, 0.0383},
		{-0.0164, 0.8985, 0.0098},
		{-0.0156, -0.0500, 0.4201},
	},
	ColorWarm: {
		{1.1, 0.05, 0.0},
		{0.0, 1.0, 0.0},
		{0.05, 0.0, 0.9},
	},
	ColorSunset: {
		{1.2, 0.2, 0.0},
		{0.1, 1.0, 0.1},
		{0.0, 0.1, 0.6},
	},
	ColorSepia: {
		{0.393, 0.769, 0.189},
		{0.349, 0.686, 0.168},
		{0.272, 0.534, 0.131},
	},
	ColorVintage: {
		{1.0, 0.6, 0.2},
		{0.3, 0.9, 0.2},
		{0.2, 0.4, 0.6},
	},
	ColorDusk: {
		{1.1, 0.0, 0.2},
		{0.0, 0.8, 0.1},
		{0.0, 0.1, 0.6},
	},
	ColorCool: {
		{0.9, 0.0, 0.2},
		{0.0, 1.0, 0.0},
		{-0.1, 0.0, 1.1},
	},
	ColorBW: {
		{0.3, 0.59, 0.11},
		{0.3, 0.59, 0.11},
		{0.3, 0.59, 0.11},
	},
	ColorIce: {
		{0.8, 0.9, 1.0},
		{0.8, 0.9, 1.0},
		{1.0, 1.0, 1.2},
	},
	ColorMoonlight: {
		{0.6, 0.2, 0.4},
		{0.2, 0.7, 0.2},
		{0.3, 0.3, 0.9},
	},
	ColorNeon: {
		{0.9, 0.0, 1.1},
		{0.0, 1.0, 0.6},
		{0.2, 0.5, 1.3},
	},
	ColorPastel: {
		{1.2, 0.1, 0.1},
		{0.1, 1.2, 0.1},
		{0.1, 0.1, 1.2},
	},
}

// MarshalText implements the encoding.TextMarshaler interface for ColorFilter
func (i ColorFilter) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for ColorFilter
func (i *ColorFilter) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*i = ColorNone
		return nil
	}
	var err error
	*i, err = ColorFilterString(string(text))
	return err
}

var ErrInvalidColorFilter = fmt.Errorf("invalid color filter")

func (i ColorFilter) ImageFilter() (ImageFilter, error) {
	if i == ColorNone {
		return nil, nil
	}
	if filter, ok := colorFilters[i]; ok {
		return ColorMatrix(filter), nil
	}
	return nil, fmt.Errorf("%w: %q", ErrInvalidColorFilter, i)
}

func (f ColorFilter) Description() (string, error) {
	switch f {
	case ColorNone:
		return "No transformation", nil
	case ColorDimmed:
		return "Darkens image uniformly while preserving hue", nil
	case ColorRedShift:
		return "CCT-derived chromatic adaptation matrix (~3400K target)", nil
	case ColorWarm:
		return "Adds a subtle warm, orange/yellow hue", nil
	case ColorSunset:
		return "Emulates deep pink/orange of a setting sun", nil
	case ColorSepia:
		return "Adds a warm, antique brown tone mimicking aged photographs", nil
	case ColorVintage:
		return "Muted, brown/green nostalgic tones", nil
	case ColorDusk:
		return "Fades brightness, adds reddish cast", nil
	case ColorCool:
		return "Adds a cool, blue tint", nil
	case ColorBW:
		return "Converts image to perceptual grayscale using luminance weights", nil
	case ColorIce:
		return "Pale desaturation with bluish cast", nil
	case ColorMoonlight:
		return "Dim blue-gray, night lighting effect", nil
	case ColorNeon:
		return "Boosts contrast, magenta-blue cyberpunk", nil
	case ColorPastel:
		return "Softens tones, gentle highlight boost", nil
	}
	return "", fmt.Errorf("%w: %q", ErrInvalidColorFilter, f)
}

type RenderFilters struct {
	Magnify     int         `json:"magnify,omitempty"`
	ColorFilter ColorFilter `json:"color_filter,omitempty"`
	Output2x    bool        `json:"2x,omitempty"`
}

func (f RenderFilters) String() string {
	return fmt.Sprintf("Magnify=%d, ColorFilter=%q, Output2x=%t", f.Magnify, f.ColorFilter, f.Output2x)
}

// Chain applies a sequence of ImageFilters in order.
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

// Magnify enlarges an image by an integer factor.
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

// ColorMatrix applies a 3x3 color transformation matrix to the RGB values of an image.
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
