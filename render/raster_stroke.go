package render

import (
	"image"
	"math"

	"github.com/tronbyt/gg"
)

func strokeBrush(width float64) []image.Point {
	radius := width / 2.0
	if radius <= 0.5 {
		return []image.Point{{X: 0, Y: 0}}
	}

	limit := int(math.Ceil(radius))
	brush := make([]image.Point, 0, (2*limit+1)*(2*limit+1))
	for y := -limit; y <= limit; y++ {
		for x := -limit; x <= limit; x++ {
			if math.Hypot(float64(x), float64(y)) <= radius {
				brush = append(brush, image.Point{X: x, Y: y})
			}
		}
	}

	if len(brush) == 0 {
		return []image.Point{{X: 0, Y: 0}}
	}

	return brush
}

func setRasterPixel(dc *gg.Context, x, y int, brush []image.Point) {
	for _, offset := range brush {
		dc.SetPixel(x+offset.X, y+offset.Y)
	}
}

func drawRasterLine(dc *gg.Context, x0, y0, x1, y1 int, brush []image.Point) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}

	sx := -1
	if x0 < x1 {
		sx = 1
	}

	dy := y1 - y0
	if dy > 0 {
		dy = -dy
	}

	sy := -1
	if y0 < y1 {
		sy = 1
	}

	err := dx + dy

	for {
		setRasterPixel(dc, x0, y0, brush)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func transformPoint(dc *gg.Context, x, y float64) image.Point {
	tx, ty := dc.TransformPoint(x, y)
	return image.Point{X: int(tx), Y: int(ty)}
}

func drawRasterizedLine(dc *gg.Context, x0, y0, x1, y1, width float64) {
	brush := strokeBrush(width)
	p0 := transformPoint(dc, x0, y0)
	p1 := transformPoint(dc, x1, y1)
	drawRasterLine(dc, p0.X, p0.Y, p1.X, p1.Y, brush)
}

func drawRasterizedArc(dc *gg.Context, x, y, radius, startAngle, endAngle, width float64) {
	sweep := endAngle - startAngle
	steps := max(int(math.Ceil(math.Abs(sweep)*math.Max(radius, 1)*2)), 1)

	brush := strokeBrush(width)
	prev := transformPoint(
		dc,
		x+radius*math.Cos(startAngle),
		y+radius*math.Sin(startAngle),
	)

	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		angle := startAngle + sweep*t
		next := transformPoint(
			dc,
			x+radius*math.Cos(angle),
			y+radius*math.Sin(angle),
		)
		drawRasterLine(dc, prev.X, prev.Y, next.X, next.Y, brush)
		prev = next
	}
}

func drawRasterizedPolygonStroke(dc *gg.Context, vertices []Point, width float64) {
	if len(vertices) < 2 {
		return
	}

	brush := strokeBrush(width)
	prev := transformPoint(dc, vertices[0].X, vertices[0].Y)
	for _, vertex := range vertices[1:] {
		next := transformPoint(dc, vertex.X, vertex.Y)
		drawRasterLine(dc, prev.X, prev.Y, next.X, next.Y, brush)
		prev = next
	}

	first := transformPoint(dc, vertices[0].X, vertices[0].Y)
	drawRasterLine(dc, prev.X, prev.Y, first.X, first.Y, brush)
}
