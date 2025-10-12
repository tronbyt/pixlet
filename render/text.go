package render

import (
	"image"
	"image/color"

	"github.com/tidbyt/gg"
	"golang.org/x/image/font"
)

var (
	DefaultFontFace  = "tb-8"
	DefaultFontColor = color.White
	MaxWidth         = 1000
)

// Text draws a string of text on a single line.
//
// By default, the text will use the "tb-8" font, but other fonts can
// be chosen via the `font` attribute. The `height` and `offset`
// parameters allow fine tuning of the vertical layout of the
// string. Take a look at the [font documentation](fonts.md) for more
// information.
//
// DOC(Content): The text string to draw
// DOC(Font): Desired font face
// DOC(Height): Limits height of the area on which text is drawn
// DOC(Offset): Shifts position of text vertically.
// DOC(Color): Desired font color
//
// EXAMPLE BEGIN
// render.Text(content="Tidbyt!", color="#099")
// EXAMPLE END
type Text struct {
	Widget
	Content string `starlark:"content,required"`
	Font    string
	Height  int
	Offset  int
	Color   color.Color

	img image.Image
}

func (t *Text) Size() (int, int) {
	return t.img.Bounds().Dx(), t.img.Bounds().Dy()
}

func (t *Text) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	dc.DrawImage(t.img, 0, 0)
}

func (t *Text) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	return image.Rect(0, 0, t.img.Bounds().Dx(), t.img.Bounds().Dy())
}

func (t *Text) Init() error {
	if t.Font == "" {
		t.Font = DefaultFontFace
	}
	face, err := GetFont(t.Font)
	if err != nil {
		return err
	}

	// Check if content contains emojis
	if hasAnyEmojiSequence(t.Content) {
		return t.initWithEmojis(face)
	}

	dc := gg.NewContext(0, 0)
	dc.SetFontFace(face)

	w, _ := dc.MeasureString(t.Content)
	width := int(w)

	// If the width of the text is longer then the max, cut off the size of the
	// image so it's not unbounded.
	if width > MaxWidth {
		width = MaxWidth
	}

	metrics := face.Metrics()
	ascent := metrics.Ascent.Floor()
	descent := metrics.Descent.Floor()

	height := ascent + descent
	if t.Height != 0 {
		height = t.Height
	}

	dc = gg.NewContext(width, height)
	dc.SetFontFace(face)
	if t.Color != nil {
		dc.SetColor(t.Color)
	} else {
		dc.SetColor(DefaultFontColor)
	}

	dc.DrawString(t.Content, 0, float64(height-descent-t.Offset))

	t.img = dc.Image()

	return nil
}

func (t *Text) initWithEmojis(face font.Face) error {
	segments := segmentEmoji(t.Content)
	if len(segments) == 0 {
		return nil
	}

	// Calculate total width and height needed
	metrics := face.Metrics()
	ascent := metrics.Ascent.Floor()
	descent := metrics.Descent.Floor()
	textHeight := ascent + descent

	// Emoji height is typically emojiCellH (10px)
	height := textHeight
	if emojiCellH > textHeight {
		height = emojiCellH
	}
	if t.Height != 0 {
		height = t.Height
	}

	// Calculate total width by measuring each segment
	totalWidth := 0
	dc := gg.NewContext(0, 0)
	dc.SetFontFace(face)

	for _, seg := range segments {
		if seg.emoji {
			totalWidth += emojiCellW // emoji width is emojiCellW (10px)
		} else {
			w, _ := dc.MeasureString(seg.text)
			totalWidth += int(w)
		}
	}

	// Limit width if needed
	if totalWidth > MaxWidth {
		totalWidth = MaxWidth
	}

	// Create the rendering context
	dc = gg.NewContext(totalWidth, height)

	// Create RGBA image to support drawing emojis with transparency
	rgba := image.NewRGBA(image.Rect(0, 0, totalWidth, height))
	dc = gg.NewContextForRGBA(rgba)
	dc.SetFontFace(face)

	if t.Color != nil {
		dc.SetColor(t.Color)
	} else {
		dc.SetColor(DefaultFontColor)
	}

	// Render each segment
	x := 0
	baselineY := height - descent - t.Offset

	for _, seg := range segments {
		if seg.emoji {
			// Draw emoji using the emoji system
			advance := drawEmojiSequence(rgba, seg.text, x, baselineY)
			x += advance
		} else {
			// Draw regular text
			dc.DrawString(seg.text, float64(x), float64(baselineY))
			w, _ := dc.MeasureString(seg.text)
			x += int(w)
		}

		// Stop if we exceed max width
		if x >= MaxWidth {
			break
		}
	}

	t.img = rgba
	return nil
}

func (t Text) FrameCount(bounds image.Rectangle) int {
	return 1
}
