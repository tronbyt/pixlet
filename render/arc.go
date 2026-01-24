package render

import (
	"image"
	"image/color"
	"math"

	"github.com/tronbyt/gg"
)

// Arc draws an arc. The arc is centered at (x, y).
//
// DOC(X): The x-coordinate of the center of the arc.
// DOC(Y): The y-coordinate of the center of the arc.
// DOC(Radius): The radius of the arc.
// DOC(StartAngle): The starting angle of the arc, in radians.
// DOC(EndAngle): The ending angle of the arc, in radians.
// DOC(Color): The color of the arc.
// DOC(Width): The width of the arc.
//
// EXAMPLE BEGIN
// render.Arc(
//
//	x = 10,
//	y = 10,
//	radius = 10,
//	start_angle = 0,
//	end_angle = 3.14 * 1.5,
//	width = 3,
//	color = "#0ff",
//
// )
// EXAMPLE END.
type Arc struct {
	Widget

	X          float64     `starlark:"x,required"`
	Y          float64     `starlark:"y,required"`
	Radius     float64     `starlark:"radius,required"`
	StartAngle float64     `starlark:"start_angle,required"`
	EndAngle   float64     `starlark:"end_angle,required"`
	Color      color.Color `starlark:"color,required"`
	Width      float64     `starlark:"width,required"`
}

func (a Arc) getBounds() (float64, float64, float64, float64) {
	// Start with endpoints
	x1 := a.X + a.Radius*math.Cos(a.StartAngle)
	y1 := a.Y + a.Radius*math.Sin(a.StartAngle)
	x2 := a.X + a.Radius*math.Cos(a.EndAngle)
	y2 := a.Y + a.Radius*math.Sin(a.EndAngle)

	minX := math.Min(x1, x2)
	maxX := math.Max(x1, x2)
	minY := math.Min(y1, y2)
	maxY := math.Max(y1, y2)

	// Check cardinal points (0, 90, 180, 270 degrees)
	// We need to normalize angles to [0, 2*pi)
	start := a.StartAngle
	end := a.EndAngle

	// If start > end, we are crossing 0 (e.g. 350 to 10 degrees)
	// But gg uses "draw from start to end". If start > end, it generally draws clockwise or "the long way"?
	// Wait, gg documentation says: "Angles are specified in radians and go clockwise."
	// Actually, standard math is counter-clockwise.
	// Let's assume standard behavior: from Start to End.
	// If Start < End, it's simple interval [Start, End].
	// If Start > End, it's [Start, 2*pi] U [0, End]. (Crossing 0).

	// Normalize angles to 0-2pi for comparison
	normalize := func(angle float64) float64 {
		angle = math.Mod(angle, 2*math.Pi)
		if angle < 0 {
			angle += 2 * math.Pi
		}
		return angle
	}

	normStart := normalize(start)
	normEnd := normalize(end)

	// If the original sweep was meant to be > 2pi (full circle), or specific winding,
	// checking just normalized values might be ambiguous.
	// But for bounding box, we just need to know if the cardinal directions are covered.

	// We check each cardinal direction: 0, pi/2, pi, 3pi/2
	cardinals := []float64{0, math.Pi / 2, math.Pi, 3 * math.Pi / 2}

	for _, angle := range cardinals {
		inArc := false
		if normStart <= normEnd {
			// Normal range
			if angle >= normStart && angle <= normEnd {
				inArc = true
			}
		} else {
			// Crossing 0
			if angle >= normStart || angle <= normEnd {
				inArc = true
			}
		}

		if inArc {
			x := a.X + a.Radius*math.Cos(angle)
			y := a.Y + a.Radius*math.Sin(angle)

			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}

	// Expand by half width (stroke width)
	halfWidth := a.Width / 2.0
	minX -= halfWidth
	maxX += halfWidth
	minY -= halfWidth
	maxY += halfWidth

	return minX, maxX, minY, maxY
}

func (a Arc) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	minX, maxX, minY, maxY := a.getBounds()
	return image.Rect(
		0,
		0,
		int(math.Ceil(maxX-minX)),
		int(math.Ceil(maxY-minY)),
	)
}

func (a Arc) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	minX, _, minY, _ := a.getBounds()

	dc.Push()
	dc.Translate(-minX, -minY)
	dc.SetColor(a.Color)
	dc.SetLineWidth(a.Width)
	dc.DrawArc(a.X, a.Y, a.Radius, a.StartAngle, a.EndAngle)
	dc.Stroke()
	dc.Pop()
}

func (a Arc) FrameCount(bounds image.Rectangle) int {
	return 1
}
