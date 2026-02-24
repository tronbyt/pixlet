package colorutil

import (
	"image/color"
	"math"
)

func ParseHSV(h, s, v float64, a uint8) color.NRGBA {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	s = max(0, min(s, 1))
	v = max(0, min(v, 1))

	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	var r, g, b float64

	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return color.NRGBA{
		R: uint8(math.Round((r + m) * 255)),
		G: uint8(math.Round((g + m) * 255)),
		B: uint8(math.Round((b + m) * 255)),
		A: a,
	}
}

func FormatHSV(c color.NRGBA) (h, s, v float64, a uint8) {
	r := float64(c.R) / 255.0
	g := float64(c.G) / 255.0
	b := float64(c.B) / 255.0

	maxVal := max(r, g, b)
	minVal := min(r, g, b)

	v = maxVal
	d := maxVal - minVal

	if maxVal != 0 {
		s = d / maxVal
	}

	if maxVal == minVal {
		h = 0
	} else {
		switch maxVal {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6.0
			}
		case g:
			h = (b-r)/d + 2.0
		case b:
			h = (r-g)/d + 4.0
		}
		h *= 60.0
	}

	return h, s, v, c.A
}
