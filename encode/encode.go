package encode

import (
	"crypto/sha256"
	"fmt"
	"image"

	"github.com/vmihailenco/msgpack/v5"

	"tidbyt.dev/pixlet/render"
)

const (
	WebPKMin                 = 0
	WebPKMax                 = 0
	DefaultScreenDelayMillis = 50
	DefaultMaxAgeSeconds     = 0 // 0 => no max age, cache forever!
)

type Screens struct {
	roots             []render.Root
	images            []image.Image
	delay             int32
	MaxAge            int32
	ShowFullAnimation bool
}

type ImageFilter func(image.Image) (image.Image, error)

func ScreensFromRoots(roots []render.Root) *Screens {
	screens := Screens{
		roots:  roots,
		delay:  DefaultScreenDelayMillis,
		MaxAge: DefaultMaxAgeSeconds,
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

func ScreensFromImages(images ...image.Image) *Screens {
	screens := Screens{
		images: images,
		delay:  DefaultScreenDelayMillis,
		MaxAge: DefaultMaxAgeSeconds,
	}
	return &screens
}

// Empty returns true if there are no render roots or images in this screen.
func (s *Screens) Empty() bool {
	return len(s.roots) == 0 && len(s.images) == 0
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

	if len(s.roots) == 0 {
		// there are no roots, so this might have been a screen created directly
		// from images. if so, consider the images in the hash.
		hashable.Images = s.images
	}

	j, err := msgpack.Marshal(hashable)
	if err != nil {
		return nil, fmt.Errorf("marshaling render tree to JSON: %w", err)
	}

	h := sha256.Sum256(j)
	return h[:], nil
}

func (s *Screens) render(filters ...ImageFilter) ([]image.Image, error) {
	if s.images == nil {
		s.images = render.PaintRoots(true, s.roots...)
	}

	if len(s.images) == 0 {
		return nil, nil
	}

	images := s.images

	if len(filters) > 0 {
		images = []image.Image{}
		for _, im := range s.images {
			for _, f := range filters {
				if f == nil {
					continue
				}
				imFiltered, err := f(im)
				if err != nil {
					return nil, err
				}
				im = imFiltered
			}
			images = append(images, im)
		}
	}

	return images, nil
}
