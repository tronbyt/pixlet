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
)

//go:embed sprites.png
var spritesPNG []byte

var (
	sheetImg *image.NRGBA
	mu       sync.Mutex
)

func Sheet() (*image.NRGBA, error) {
	mu.Lock()
	defer mu.Unlock()

	if sheetImg == nil {
		img, err := png.Decode(bytes.NewReader(spritesPNG))
		if err != nil {
			return nil, err
		}
		if nrgba, ok := img.(*image.NRGBA); ok {
			sheetImg = nrgba
		} else {
			rb := image.NewNRGBA(img.Bounds())
			draw.Draw(rb, rb.Bounds(), img, image.Point{}, draw.Src)
			sheetImg = rb
		}
	}

	return sheetImg, nil
}

func Get(s string) (*image.NRGBA, error) {
	// Check if the emoji exists in our index
	point, exists := Index[s]
	if !exists {
		return nil, fmt.Errorf("emoji %q not found in emoji index", s)
	}

	// Get the emoji sprite sheet
	sheet, err := Sheet()
	if err != nil {
		return nil, fmt.Errorf("failed to load emoji sheet: %w", err)
	}

	// Extract the emoji from the sprite sheet
	srcRect := image.Rect(
		point.X*CellW, point.Y*CellH,
		(point.X+1)*CellW, (point.Y+1)*CellH,
	)

	// Create source image for this emoji
	srcImg := image.NewNRGBA(image.Rect(0, 0, CellW, CellH))
	draw.Draw(srcImg, srcImg.Bounds(), sheet, srcRect.Min, draw.Src)
	return srcImg, nil
}

type Segment struct {
	Text    string // either a plain text run or the exact emoji sequence string
	IsEmoji bool
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
		if _, ok := Index[cluster]; ok {
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
