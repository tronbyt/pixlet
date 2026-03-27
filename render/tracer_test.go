package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTracerCircularPath(t *testing.T) {
	ic := ImageChecker{
		Palette: map[string]color.RGBA{
			"1": {0xff, 0xff, 0xff, 0xff},
			".": {0, 0, 0, 0},
		},
	}

	tr := Tracer{
		Path:        &CircularPath{Radius: 4},
		TraceLength: 0,
	}

	// First quadrant
	require.NoError(t, ic.Check([]string{
		"........",
		"........",
		"........",
		"........",
		".......1",
		"........",
		"........",
		"........",
	}, PaintWidget(tr, image.Rect(0, 0, 100, 100), 0)))

	require.NoError(t, ic.Check([]string{
		"........",
		"........",
		"........",
		"........",
		"........",
		".......1",
		"........",
		"........",
	}, PaintWidget(tr, image.Rect(0, 0, 100, 100), 1)))

	require.NoError(t, ic.Check([]string{
		"........",
		"........",
		"........",
		"........",
		"........",
		"........",
		".......1",
		"........",
	}, PaintWidget(tr, image.Rect(0, 0, 100, 100), 2)))

	require.NoError(t, ic.Check([]string{
		"........",
		"........",
		"........",
		"........",
		"........",
		"........",
		"........",
		"......1.",
	}, PaintWidget(tr, image.Rect(0, 0, 100, 100), 3)))

	// Spot check third quadrant
	require.NoError(t, ic.Check([]string{
		"........",
		"1.......",
		"........",
		"........",
		"........",
		"........",
		"........",
		"........",
	}, PaintWidget(tr, image.Rect(0, 0, 100, 100), 14)))

	// Last pixel and verify it loops
	require.NoError(t, ic.Check([]string{
		"........",
		"........",
		"........",
		".......1",
		"........",
		"........",
		"........",
		"........",
	}, PaintWidget(tr, image.Rect(0, 0, 100, 100), 23)))

	require.NoError(t, ic.Check([]string{
		"........",
		"........",
		"........",
		"........",
		"........",
		".......1",
		"........",
		"........",
	}, PaintWidget(tr, image.Rect(0, 0, 100, 100), 25)))

	// All in all, we should have 24 frames
	assert.Equal(t, 24, tr.FrameCount(image.Rect(0, 0, 0, 0)))
}
