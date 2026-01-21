package render

import (
	"image"
	"image/color"
	"math"

	"github.com/tronbyt/gg"
)

type Point struct {
	X, Y float64
}

// Polygon draws a polygon.
//
// DOC(Vertices): A list of (x, y) tuples representing the vertices of the polygon.
// DOC(FillColor): The color used to fill the polygon.
// DOC(StrokeColor): The color used to draw the polygon's stroke.
// DOC(StrokeWidth): The width of the polygon's stroke.
//
// EXAMPLE BEGIN
// render.Polygon(
//
//	vertices = [(0, 0), (20, 0), (20, 10), (0, 10)],
//	fill_color = "#00f",
//	stroke_color = "#fff",
//	stroke_width = 1,
//
// )
// EXAMPLE END
type Polygon struct {
	Widget
	Vertices    []Point     `starlark:"vertices,required"`
	FillColor   color.Color `starlark:"fill_color"`
	StrokeColor color.Color `starlark:"stroke_color"`
	StrokeWidth float64     `starlark:"stroke_width"`
}

func (p Polygon) getBounds() (minX, maxX, minY, maxY float64) {
	minX, minY = math.Inf(1), math.Inf(1)
	maxX, maxY = math.Inf(-1), math.Inf(-1)

	for _, pt := range p.Vertices {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}

	if p.StrokeColor != nil && p.StrokeWidth > 0 {
		halfWidth := p.StrokeWidth / 2.0
		minX -= halfWidth
		maxX += halfWidth
		minY -= halfWidth
		maxY += halfWidth
	}

	return
}

func (p Polygon) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	minX, maxX, minY, maxY := p.getBounds()

	if math.IsInf(minX, 0) {
		return image.Rect(0, 0, 0, 0)
	}

	return image.Rect(0, 0, int(math.Ceil(maxX-minX)), int(math.Ceil(maxY-minY)))
}

func (p Polygon) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	if len(p.Vertices) == 0 {
		return
	}

	minX, _, minY, _ := p.getBounds()

	dc.Push()
	dc.Translate(-minX, -minY)

	for i, pt := range p.Vertices {
		if i == 0 {
			dc.MoveTo(pt.X, pt.Y)
		} else {
			dc.LineTo(pt.X, pt.Y)
		}
	}
	dc.ClosePath()

	if p.FillColor != nil {
		dc.SetColor(p.FillColor)
		if p.StrokeColor != nil && p.StrokeWidth > 0 {
			dc.FillPreserve()
		} else {
			dc.Fill()
		}
	}

	if p.StrokeColor != nil && p.StrokeWidth > 0 {
		dc.SetColor(p.StrokeColor)
		dc.SetLineWidth(p.StrokeWidth)
		dc.Stroke()
	}

	dc.Pop()
}

func (p Polygon) FrameCount(bounds image.Rectangle) int {
	return 1
}
