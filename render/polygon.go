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
// DOC(Color): The color of the polygon.
type Polygon struct {
	Widget
	Vertices []Point     `starlark:"vertices,required"`
	Color    color.Color `starlark:"color,required"`
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
	dc.SetColor(p.Color)

	for i, pt := range p.Vertices {
		if i == 0 {
			dc.MoveTo(pt.X, pt.Y)
		} else {
			dc.LineTo(pt.X, pt.Y)
		}
	}

	dc.ClosePath()
	dc.Fill()
	dc.Pop()
}

func (p Polygon) FrameCount(bounds image.Rectangle) int {
	return 1
}
