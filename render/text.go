package render

import (
	"image"
	"image/color"

	"github.com/tronbyt/gg"
	"github.com/tronbyt/pixlet/render/emoji"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"github.com/tronbyt/pixlet/tools/i18n"
	"go.starlark.net/starlark"
)

const (
	DefaultFontFace   = "tb-8"
	DefaultFontFace2x = "terminus-16"
	MaxWidth          = 1000
)

var DefaultFontColor = color.White

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
//
//	render.Text(content="Tidbyt!", color="#099")
//
// EXAMPLE END.
type Text struct {
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

func (t *Text) Init(thread *starlark.Thread) error {
	if t.Font == "" {
		t.Font = getDefaultFont(thread)
	}

	face, err := GetFont(t.Font)
	if err != nil {
		return err
	}

	content := i18n.VisualBidiString(t.Content)

	// Check if content contains emojis
	segments, hasEmoji := emoji.SegmentString(content)

	dc := gg.NewContext(0, 0)
	dc.SetFontFace(face)

	var width int
	for _, seg := range segments {
		width += seg.Width(dc)
	}

	// If the width of the text is longer than the max, cut off the size of the
	// image so it's not unbounded.
	if width > MaxWidth {
		width = MaxWidth
	}

	metrics := face.Metrics()
	ascent := metrics.Ascent.Floor()
	descent := metrics.Descent.Floor()

	height := ascent + descent
	if hasEmoji && emoji.MaxHeight > height {
		height = emoji.MaxHeight
	}
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

	// Render each segment
	x, y := 0, height-descent-t.Offset
	for _, seg := range segments {
		x += seg.Draw(dc, x, y)

		// Stop if we exceed max width
		if x >= MaxWidth {
			break
		}
	}

	t.img = dc.Image()
	return nil
}

func (t Text) FrameCount(bounds image.Rectangle) int {
	return 1
}

func getDefaultFont(thread *starlark.Thread) string {
	if meta, err := canvas.FromThread(thread); err == nil {
		if meta.Is2x {
			return DefaultFontFace2x
		}
	}
	return DefaultFontFace
}
