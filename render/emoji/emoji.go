package emoji

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"strings"
	"sync"

	"github.com/rivo/uniseg"
	"github.com/tidbyt/gg"
	font "tidbyt.dev/pixlet/fonts/emoji"
)

const (
	MaxHeight = font.MaxHeight
	MaxWidth  = font.MaxWidth
)

var (
	sheetImg  *image.NRGBA
	sheetOnce sync.Once
)

func Sheet() (*image.NRGBA, error) {
	var err error

	sheetOnce.Do(func() {
		var img image.Image
		if img, err = png.Decode(bytes.NewReader(font.Sprites)); err != nil {
			return
		}

		if nrgba, ok := img.(*image.NRGBA); ok {
			sheetImg = nrgba
		} else {
			rb := image.NewNRGBA(img.Bounds())
			draw.Draw(rb, rb.Bounds(), img, image.Point{}, draw.Src)
			sheetImg = rb
		}
	})

	return sheetImg, err
}

const variationSequence = "\uFE0F"

func Get(s string, tryVariation bool) (*image.NRGBA, error) {
	// Check if the emoji exists in our index
	bounds, exists := font.Index[s]
	if !exists {
		if tryVariation {
			bounds, exists = font.Index[s+variationSequence]
		}
		if !exists {
			return nil, fmt.Errorf("emoji %q not found in emoji index", s)
		}
	}

	// Get the emoji sprite sheet
	sheet, err := Sheet()
	if err != nil {
		return nil, fmt.Errorf("failed to load emoji sheet: %w", err)
	}

	// Create source image for this emoji with padding applied horizontally.
	srcImg := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(srcImg, srcImg.Bounds(), sheet, bounds.Min, draw.Src)
	return srcImg, nil
}

type Segment struct {
	Text    string // either a plain text run or the exact emoji sequence string
	IsEmoji bool
}

func (s Segment) Width(dc *gg.Context) int {
	if s.IsEmoji {
		if glyph, ok := font.Index[s.Text]; ok {
			return glyph.Dx()
		}
		return font.MaxWidth
	} else {
		w, _ := dc.MeasureString(s.Text)
		return int(w)
	}
}

func (s Segment) Draw(dc *gg.Context, x, y int) int {
	if s.IsEmoji {
		if srcImg, err := Get(s.Text, false); err == nil {
			dc.DrawImage(srcImg, x, y-srcImg.Bounds().Dy())
			return srcImg.Bounds().Dx()
		}
	} else {
		dc.DrawString(s.Text, float64(x), float64(y))
		w, _ := dc.MeasureString(s.Text)
		return int(w)
	}
	return 0
}

// SegmentString breaks a string into a sequence of tokens, where each token is either
// an emoji sequence key present in Index, or a plain Text segment (no emoji inside).
// Segments are identified using Unicode grapheme clusters so complex emoji remain intact.
func SegmentString(s string) ([]Segment, bool) {
	var hasEmoji bool
	segments := make([]Segment, 0, 1)
	var buf strings.Builder
	buf.Grow(len(s))

	state := -1
	for len(s) != 0 {
		var cluster string
		cluster, s, _, state = uniseg.FirstGraphemeClusterInString(s, state)
		if _, ok := font.Index[cluster]; ok {
			if buf.Len() != 0 {
				segments = append(segments, Segment{Text: buf.String()})
				buf.Reset()
			}
			hasEmoji = true
			segments = append(segments, Segment{
				IsEmoji: true,
				Text:    cluster,
			})
		} else {
			buf.WriteString(cluster)
		}
	}

	if buf.Len() != 0 {
		segments = append(segments, Segment{Text: buf.String()})
	}

	return segments, hasEmoji
}
