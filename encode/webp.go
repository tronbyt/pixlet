package encode

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/tronbyt/go-libwebp/webp"
)

const (
	WebPLevelDefault = int32(6)
	webpLevelEnv     = "PIXLET_WEBP_LEVEL"
)

var webpLevel atomic.Int32

func init() {
	if raw := os.Getenv(webpLevelEnv); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err == nil {
			SetWebPLevel(int32(parsed))
			return
		}
		slog.Warn(webpLevelEnv+" is invalid; using default.", "error", err)
	}

	webpLevel.Store(WebPLevelDefault)
}

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

	remainingDuration := time.Duration(maxDuration) * time.Millisecond
	level := int(webpLevel.Load())
	config, err := webp.ConfigLosslessPreset(level)
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

func SetWebPLevel(level int32) {
	if level < 0 || level > 9 {
		slog.Warn("WebP level is out of range (0-9); using default.", "value", level)
		return
	}
	webpLevel.Store(level)
}
