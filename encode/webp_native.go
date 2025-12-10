//go:build nativewebp

package encode

import (
	"bytes"
	"fmt"
	"image"
	"time"

	"github.com/HugoSmits86/nativewebp"
)

// Renders a screen to WebP. Optionally pass filters for
// postprocessing each individual frame.
func (s *Screens) EncodeWebP(maxDuration time.Duration, filters ...ImageFilter) ([]byte, error) {
	images, err := s.render(filters...)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return []byte{}, nil
	}

	frames := make([]image.Image, 0, len(images))
	durations := make([]uint, 0, len(images))
	disposals := make([]uint, 0, len(images))

	remainingDuration := maxDuration

	for _, im := range images {
		frameDuration := time.Duration(s.delay) * time.Millisecond

		if maxDuration > 0 {
			if frameDuration > remainingDuration {
				frameDuration = remainingDuration
			}
			remainingDuration -= frameDuration
		}

		frames = append(frames, im)
		durations = append(durations, uint(frameDuration.Milliseconds()))
		disposals = append(disposals, 0) // 0 = Unspecified

		if maxDuration > 0 && remainingDuration <= 0 {
			break
		}
	}

	buf := new(bytes.Buffer)
	if len(frames) == 1 {
		if err := nativewebp.Encode(buf, frames[0], nil); err != nil {
			return nil, fmt.Errorf("%s: %w", "encoding image", err)
		}
	} else {
		anim := nativewebp.Animation{
			Images:          frames,
			Durations:       durations,
			Disposals:       disposals,
			LoopCount:       0,
			BackgroundColor: 0x00000000,
		}

		if err := nativewebp.EncodeAll(buf, &anim, nil); err != nil {
			return nil, fmt.Errorf("%s: %w", "encoding animation", err)
		}
	}

	return buf.Bytes(), nil
}
