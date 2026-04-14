package encode

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/tronbyt/go-libwebp/webp"
	"go.uber.org/atomic"
)

const (
	webpLevelEnv     = "PIXLET_WEBP_LEVEL"
	WebPLevelDefault = 6
)

var webpLevel = atomic.NewInt32(WebPLevelDefault)

func init() { //nolint:gochecknoinits
	if raw := os.Getenv(webpLevelEnv); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err == nil {
			SetWebPLevel(int32(parsed))
			return
		}
		slog.Warn(webpLevelEnv+" is invalid; using default.", "error", err)
	}
}

// EncodeWebP renders a screen to WebP. Optionally pass filters for
// postprocessing each individual frame.
func (s *Screens) EncodeWebP(ctx context.Context, maxDuration time.Duration, filters ...ImageFilter) ([]byte, error) {
	level := int(webpLevel.Load())
	remainingDuration := maxDuration

	config, err := webp.ConfigLosslessPreset(level)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "configuring encoder", err)
	}

	var anim *webp.AnimationEncoder
	defer func() {
		if anim != nil {
			anim.Close()
		}
	}()

	var frameCount int
	for im := range s.render(ctx, filters...) {
		if anim == nil {
			bounds := im.Bounds()
			if anim, err = webp.NewAnimationEncoder(
				bounds.Dx(),
				bounds.Dy(),
				WebPKMin,
				WebPKMax,
			); err != nil {
				return nil, fmt.Errorf("%s: %w", "initializing encoder", err)
			}
		}

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
		frameCount++

		if maxDuration > 0 && remainingDuration <= 0 {
			break
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if frameCount == 0 {
		return []byte{}, nil
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
