package render

import (
	"fmt"
	"image"
	"math"

	"github.com/nfnt/resize"
	"github.com/tidbyt/gg"
	"tidbyt.dev/pixlet/render/emoji"
)

// Emoji renders a single emoji at a specified height, maintaining aspect ratio.
// This allows for rendering emojis much larger than the standard 10x10 pixel size
// used in text rendering.
//
// DOC(Emoji): The Unicode emoji sequence to render
// DOC(Width): Scale emoji to this width
// DOC(Height): Scale emoji to this height
//
// EXAMPLE BEGIN
// render.Emoji(emoji="😀", height=32) // Large smiley face
// EXAMPLE END
type Emoji struct {
	Widget
	EmojiStr      string `starlark:"emoji,required"`
	Width, Height int

	img image.Image
}

func (e *Emoji) Size() (int, int) {
	if e.img == nil {
		return 0, 0
	}
	return e.img.Bounds().Dx(), e.img.Bounds().Dy()
}

func (e *Emoji) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	if e.img != nil {
		dc.DrawImage(e.img, 0, 0)
	}
}

func (e *Emoji) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	if e.img == nil {
		return image.Rect(0, 0, 0, 0)
	}
	return image.Rect(0, 0, e.img.Bounds().Dx(), e.img.Bounds().Dy())
}

func (e *Emoji) Init() error {
	if e.Height < 0 {
		return fmt.Errorf("emoji height must not be negative, got %d", e.Height)
	}

	if e.EmojiStr == "" {
		return fmt.Errorf("emoji string cannot be empty")
	}

	srcImg, err := emoji.Get(e.EmojiStr, true)
	if err != nil {
		return fmt.Errorf("failed to get emoji: %w", err)
	}

	w, h := srcImg.Bounds().Dx(), srcImg.Bounds().Dy()

	nw, nh := e.Width, e.Height
	if nw == 0 {
		nw = int(math.Round(float64(nh) * float64(w) / float64(h)))
	}
	if nh == 0 {
		nh = int(math.Round(float64(nw) * float64(h) / float64(w)))
	}

	// Fast path: exact integer scaling
	if nw%w == 0 && nh%h == 0 {
		e.img = resize.Resize(uint(nw), uint(nh), srcImg, resize.NearestNeighbor)
		return nil
	}

	// Compute the desired scale and choose the smallest integer >= it.
	sx := float64(nw) / float64(w)
	sy := float64(nh) / float64(h)
	upFactor := int(math.Ceil(math.Max(sx, sy)))
	if upFactor < 2 {
		upFactor = 2 // oversample a bit to improve output quality
	}

	// Cap to avoid large intermediates
	const maxFactor = 10
	if upFactor > maxFactor {
		upFactor = maxFactor
	}

	// Step 1: integer upscale (nearest) to preserve pixel edges
	upW, upH := w*upFactor, h*upFactor
	up := resize.Resize(uint(upW), uint(upH), srcImg, resize.NearestNeighbor)

	// Step 2: downscale to final with a smooth filter
	e.img = resize.Resize(uint(nw), uint(nh), up, resize.Lanczos2)
	return nil
}

func (e Emoji) FrameCount(bounds image.Rectangle) int {
	return 1
}
