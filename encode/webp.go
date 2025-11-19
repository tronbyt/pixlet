package encode

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/tronbyt/go-libwebp/webp"
)

const (
	CompressionLevelEnv     = "PIXLET_WEBP_COMPRESSION_LEVEL"
	DefaultCompressionLevel = 6
)

// Renders a screen to WebP. Optionally pass filters for
// postprocessing each individual frame.
func (s *Screens) EncodeWebP(maxDuration int, filters ...ImageFilter) ([]byte, error) {
	images, err := s.render(filters...)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return []byte{}, nil
	}

	bounds := images[0].Bounds()
	anim, err := webp.NewAnimationEncoder(
		bounds.Dx(),
		bounds.Dy(),
		WebPKMin,
		WebPKMax,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "initializing encoder", err)
	}
	defer anim.Close()

	compressionLevel := DefaultCompressionLevel
	if raw := os.Getenv(CompressionLevelEnv); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 32); err == nil {
			if parsed < 0 || parsed > 9 {
				slog.Warn(CompressionLevelEnv+" is out of range (0-9); using default", "value", parsed)
			} else {
				compressionLevel = int(parsed)
			}
		} else {
			slog.Warn(CompressionLevelEnv+" is invalid; using default", "error", err)
		}
	}

	remainingDuration := time.Duration(maxDuration) * time.Millisecond
	config, err := webp.ConfigLosslessPreset(compressionLevel)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "configuring encoder", err)
	}
	for _, im := range images {
		frameDuration := time.Duration(s.delay) * time.Millisecond

		if maxDuration > 0 {
			if frameDuration > remainingDuration {
				frameDuration = remainingDuration
			}
			remainingDuration -= frameDuration
		}

		if err := anim.AddFrame(im, frameDuration, config); err != nil {
			return nil, fmt.Errorf("%s: %w", "adding frame", err)
		}

		if maxDuration > 0 && remainingDuration <= 0 {
			break
		}
	}

	buf, err := anim.Assemble()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "encoding animation", err)
	}

	return buf, nil
}
