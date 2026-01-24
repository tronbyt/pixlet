package encode

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"time"

	"github.com/ericpauley/go-quantize/quantize"
)

// EncodeGIF renders a screen to GIF. Optionally pass filters for postprocessing
// each individual frame.
func (s *Screens) EncodeGIF(maxDuration time.Duration, filters ...ImageFilter) ([]byte, error) {
	images, err := s.render(filters...)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return []byte{}, nil
	}

	g := &gif.GIF{}

	remainingDuration := maxDuration
	for imIdx, im := range images {
		imRGBA, ok := im.(*image.RGBA)
		if !ok {
			return nil, fmt.Errorf("image %d is %T, require RGBA", imIdx, im)
		}

		palette := quantize.MedianCutQuantizer{}.Quantize(make([]color.Color, 0, 256), im)
		imPaletted := image.NewPaletted(imRGBA.Bounds(), palette)
		draw.Draw(imPaletted, imRGBA.Bounds(), imRGBA, image.Point{0, 0}, draw.Src)

		frameDelay := time.Duration(s.delay) * time.Millisecond
		if maxDuration > 0 {
			if frameDelay > remainingDuration {
				frameDelay = remainingDuration
			}
			remainingDuration -= frameDelay
		}

		g.Image = append(g.Image, imPaletted)
		g.Delay = append(g.Delay, int(frameDelay.Milliseconds()/10)) // in 100ths of a second

		if maxDuration > 0 && remainingDuration <= 0 {
			break
		}
	}

	buf := &bytes.Buffer{}
	err = gif.EncodeAll(buf, g)
	if err != nil {
		return nil, fmt.Errorf("encoding: %w", err)
	}

	return buf.Bytes(), nil
}
