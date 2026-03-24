package encode

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"iter"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/tronbyt/pixlet/render"
)

const (
	WebPKMin                 = 0
	WebPKMax                 = 0
	DefaultScreenDelayMillis = 50
	DefaultMaxAgeSeconds     = 0 // 0 => no max age, cache forever!
)

type Screens struct {
	roots             []render.Root
	delay             int32
	MaxAge            int32
	ShowFullAnimation bool
	width             int
	height            int
}

type ImageFilter func(image.Image) image.Image

func ScreensFromRoots(roots []render.Root, width int, height int) *Screens {
	screens := Screens{
		roots:  roots,
		delay:  DefaultScreenDelayMillis,
		MaxAge: DefaultMaxAgeSeconds,
		width:  width,
		height: height,
	}
	if len(roots) > 0 {
		if roots[0].Delay > 0 {
			screens.delay = roots[0].Delay
		}
		if roots[0].MaxAge > 0 {
			screens.MaxAge = roots[0].MaxAge
		}
		screens.ShowFullAnimation = roots[0].ShowFullAnimation
	}
	return &screens
}

// Empty returns true if there are no render roots or images in this screen.
func (s *Screens) Empty() bool {
	return len(s.roots) == 0
}

// Hash returns a hash of the render roots for this screen. This can be used for
// testing whether two render trees are exactly equivalent, without having to
// do the actual rendering.
func (s *Screens) Hash() ([]byte, error) {
	hashable := struct {
		Roots  []render.Root
		Images []image.Image
		Delay  int32
		MaxAge int32
	}{
		Roots:  s.roots,
		Delay:  s.delay,
		MaxAge: s.MaxAge,
	}

	j, err := msgpack.Marshal(hashable)
	if err != nil {
		return nil, fmt.Errorf("marshaling render tree to JSON: %w", err)
	}

	h := sha256.Sum256(j)
	return h[:], nil
}

func (s *Screens) render(ctx context.Context, filters ...ImageFilter) iter.Seq[image.Image] {
	return func(yield func(image.Image) bool) {
		for i := range s.roots {
			for img := range s.roots[i].Paint(ctx, s.width, s.height, true) {
				for _, fn := range filters {
					if fn != nil {
						img = fn(img)
					}
				}

				if !yield(img) {
					return
				}
			}
		}
	}
}
