package render

import (
	"fmt"
	"image"

	"github.com/tidbyt/gg"
	"tidbyt.dev/pixlet/render/emoji"
)

// Emoji renders a single emoji at a specified height, maintaining aspect ratio.
// This allows for rendering emojis much larger than the standard 10x10 pixel size
// used in text rendering.
//
// DOC(Emoji): The Unicode emoji sequence to render
// DOC(Height): Desired height in pixels (width will be calculated to maintain aspect ratio)
//
// EXAMPLE BEGIN
// render.Emoji(emoji="😀", height=32)  // Large smiley face
// EXAMPLE END
type Emoji struct {
	Widget
	EmojiStr string `starlark:"emoji,required"`
	Height   int    `starlark:"height,required"`

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
	if e.Height <= 0 {
		return fmt.Errorf("emoji height must be positive, got %d", e.Height)
	}

	if e.EmojiStr == "" {
		return fmt.Errorf("emoji string cannot be empty")
	}

	srcImg, err := emoji.Get(e.EmojiStr)
	if err != nil {
		return fmt.Errorf("failed to get emoji: %w", err)
	}

	// Calculate scaled dimensions (maintaining aspect ratio)
	// Emojis are square (CellW == CellH), so width = height
	scaledWidth := e.Height
	scaledHeight := e.Height

	// Create the scaled image using gg for high-quality scaling
	dc := gg.NewContext(scaledWidth, scaledHeight)

	// Scale and draw the emoji
	scaleX := float64(scaledWidth) / float64(srcImg.Bounds().Dx())
	scaleY := float64(scaledHeight) / float64(srcImg.Bounds().Dy())

	dc.Scale(scaleX, scaleY)
	dc.DrawImage(srcImg, 0, 0)

	e.img = dc.Image()
	return nil
}

func (e Emoji) FrameCount(bounds image.Rectangle) int {
	return 1
}
