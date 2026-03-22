package render

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tronbyt/gg"
)

func TestLineDefaultsToBinaryStroke(t *testing.T) {
	im := PaintWidget(Line{
		X1:    0,
		Y1:    0,
		X2:    15,
		Y2:    7,
		Width: 1,
		Color: color.RGBA{0xff, 0xff, 0xff, 0xff},
	}, image.Rect(0, 0, 16, 8), 0)

	assertBinaryAlpha(t, im)
	assertHasPaintedPixels(t, im)
}

func TestLineAntiAliasOptIn(t *testing.T) {
	im := PaintWidget(Line{
		X1:        0,
		Y1:        0,
		X2:        15,
		Y2:        7,
		Width:     1,
		Color:     color.RGBA{0xff, 0xff, 0xff, 0xff},
		AntiAlias: true,
	}, image.Rect(0, 0, 16, 8), 0)

	assertHasIntermediateAlpha(t, im)
}

func TestLineNoAADoesNotDropTransforms(t *testing.T) {
	dc := gg.NewContext(20, 20)
	dc.Translate(10, 5)

	Line{
		X1:    0,
		Y1:    0,
		X2:    1,
		Y2:    0,
		Width: 1,
		Color: color.RGBA{0xff, 0xff, 0xff, 0xff},
	}.Paint(dc, image.Rect(0, 0, 20, 20), 0)

	assert.Equal(t, color.RGBA{}, color.RGBAModel.Convert(dc.Image().At(0, 0)))
	assert.Equal(t, color.RGBA{0xff, 0xff, 0xff, 0xff}, color.RGBAModel.Convert(dc.Image().At(10, 5)))
	assert.Equal(t, color.RGBA{0xff, 0xff, 0xff, 0xff}, color.RGBAModel.Convert(dc.Image().At(11, 5)))
}

func TestArcDefaultsToBinaryStroke(t *testing.T) {
	im := PaintWidget(Arc{
		X:          8,
		Y:          8,
		Radius:     6,
		StartAngle: 0,
		EndAngle:   math.Pi * 1.5,
		Width:      1,
		Color:      color.RGBA{0xff, 0xff, 0xff, 0xff},
	}, image.Rect(0, 0, 16, 16), 0)

	assertBinaryAlpha(t, im)
	assertHasPaintedPixels(t, im)
}

func TestArcAntiAliasOptIn(t *testing.T) {
	im := PaintWidget(Arc{
		X:          8,
		Y:          8,
		Radius:     6,
		StartAngle: 0,
		EndAngle:   math.Pi * 1.5,
		Width:      1,
		Color:      color.RGBA{0xff, 0xff, 0xff, 0xff},
		AntiAlias:  true,
	}, image.Rect(0, 0, 16, 16), 0)

	assertHasIntermediateAlpha(t, im)
}

func TestPolygonDefaultsToBinaryStroke(t *testing.T) {
	im := PaintWidget(Polygon{
		Vertices: []Point{
			{X: 1, Y: 1},
			{X: 14, Y: 2},
			{X: 8, Y: 13},
		},
		StrokeColor: color.RGBA{0xff, 0xff, 0xff, 0xff},
		StrokeWidth: 1,
	}, image.Rect(0, 0, 16, 16), 0)

	assertBinaryAlpha(t, im)
	assertHasPaintedPixels(t, im)
}

func TestPolygonAntiAliasOptIn(t *testing.T) {
	im := PaintWidget(Polygon{
		Vertices: []Point{
			{X: 1, Y: 1},
			{X: 14, Y: 2},
			{X: 8, Y: 13},
		},
		StrokeColor: color.RGBA{0xff, 0xff, 0xff, 0xff},
		StrokeWidth: 1,
		AntiAlias:   true,
	}, image.Rect(0, 0, 16, 16), 0)

	assertHasIntermediateAlpha(t, im)
}

func assertBinaryAlpha(t *testing.T, im image.Image) {
	t.Helper()
	bounds := im.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			a := color.RGBAModel.Convert(im.At(x, y)).(color.RGBA).A
			assert.Contains(t, []uint8{0x00, 0xff}, a, "expected binary alpha at %d,%d", x, y)
		}
	}
}

func assertHasIntermediateAlpha(t *testing.T, im image.Image) {
	t.Helper()
	bounds := im.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			a := color.RGBAModel.Convert(im.At(x, y)).(color.RGBA).A
			if a != 0x00 && a != 0xff {
				return
			}
		}
	}
	t.Fatalf("expected at least one partially transparent pixel")
}

func assertHasPaintedPixels(t *testing.T, im image.Image) {
	t.Helper()
	bounds := im.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.RGBAModel.Convert(im.At(x, y)).(color.RGBA).A != 0 {
				return
			}
		}
	}
	t.Fatalf("expected at least one painted pixel")
}
