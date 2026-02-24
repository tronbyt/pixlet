package colorutil

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

var ErrInvalidLength = fmt.Errorf("invalid length")

func ParseHex(text string) (color.NRGBA, error) {
	var c color.NRGBA
	trimmed := strings.TrimPrefix(text, "#")

	switch len(trimmed) {
	case 3, 4:
		parsed, err := strconv.ParseUint(trimmed, 16, 16)
		if err != nil {
			return c, err
		}

		if len(trimmed) == 4 {
			c.A = uint8(parsed & 0xF)
			c.A |= c.A << 4
			parsed >>= 4
		} else {
			c.A = 0xFF
		}
		c.B = uint8(parsed & 0xF)
		c.B |= c.B << 4
		parsed >>= 4
		c.G = uint8(parsed & 0xF)
		c.G |= c.G << 4
		parsed >>= 4
		c.R = uint8(parsed & 0xF)
		c.R |= c.R << 4
	case 6, 8:
		parsed, err := strconv.ParseUint(trimmed, 16, 32)
		if err != nil {
			return c, err
		}

		if len(trimmed) == 8 {
			c.A = uint8(parsed & 0xFF)
			parsed >>= 8
		} else {
			c.A = 0xFF
		}
		c.B = uint8(parsed & 0xFF)
		parsed >>= 8
		c.G = uint8(parsed & 0xFF)
		parsed >>= 8
		c.R = uint8(parsed & 0xFF)
	default:
		return c, ErrInvalidLength
	}

	return c, nil
}

func FormatHex(c color.NRGBA) string {
	skipAlpha := c.A == 0xFF
	rgbShorthand := c.R>>4 == c.R&0xF && c.G>>4 == c.G&0xF && c.B>>4 == c.B&0xF
	aShorthand := c.A>>4 == c.A&0xF

	if rgbShorthand {
		if skipAlpha {
			return fmt.Sprintf("#%x%x%x", c.R&0xF, c.G&0xF, c.B&0xF)
		}
		if aShorthand {
			return fmt.Sprintf("#%x%x%x%x", c.R&0xF, c.G&0xF, c.B&0xF, c.A&0xF)
		}
	}

	if skipAlpha {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
}
