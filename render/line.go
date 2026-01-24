package render

import (
	"image"
	"image/color"
	"math"

	"github.com/tronbyt/gg"
)

// Line draws a line from (x1, y1) to (x2, y2).
//
// DOC(X1): The x-coordinate of the starting point.
// DOC(Y1): The y-coordinate of the starting point.
// DOC(X2): The x-coordinate of the ending point.
// DOC(Y2): The y-coordinate of the ending point.
// DOC(Color): The color of the line.
// DOC(Width): The width of the line.
//
// EXAMPLE BEGIN
// render.Line(
//
//	x1 = 0,
//	y1 = 0,
//	x2 = 63,
//	y2 = 31,
//	width = 1,
//	color = "#fff",
//
// )
// EXAMPLE END.
type Line struct {
	Widget

	X1    float64     `starlark:"x1,required"`
	Y1    float64     `starlark:"y1,required"`
	X2    float64     `starlark:"x2,required"`
	Y2    float64     `starlark:"y2,required"`
	Color color.Color `starlark:"color,required"`
	Width float64     `starlark:"width,required"`
}

func (l Line) getBounds() (float64, float64, float64, float64) {
	minX := math.Min(l.X1, l.X2)
	maxX := math.Max(l.X1, l.X2)
	minY := math.Min(l.Y1, l.Y2)
	maxY := math.Max(l.Y1, l.Y2)

	halfWidth := l.Width / 2.0

	// Ensure the bounds encompass the stroke width
	minX -= halfWidth
	maxX += halfWidth
	minY -= halfWidth
	maxY += halfWidth

	return minX, maxX, minY, maxY
}

func (l Line) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	minX, maxX, minY, maxY := l.getBounds()
	return image.Rect(0, 0, int(math.Ceil(maxX-minX)), int(math.Ceil(maxY-minY)))
}

func (l Line) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	minX, _, minY, _ := l.getBounds()

	dc.Push()
	dc.Translate(-minX, -minY)
	dc.SetColor(l.Color)
	dc.SetLineWidth(l.Width)
	dc.DrawLine(l.X1, l.Y1, l.X2, l.Y2)
	dc.Stroke()
	dc.Pop()
}

func (l Line) FrameCount(bounds image.Rectangle) int {
	return 1
}
