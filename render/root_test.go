package render

import (
	"context"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tronbyt/gg"
)

// panickingWidget panics when painting the frame with index panicAt.
type panickingWidget struct {
	frames  int
	panicAt int
}

func (p panickingWidget) FrameCount(bounds image.Rectangle) int {
	return p.frames
}

func (p panickingWidget) PaintBounds(bounds image.Rectangle, frameIdx int) image.Rectangle {
	return bounds
}

func (p panickingWidget) Paint(dc *gg.Context, bounds image.Rectangle, frameIdx int) {
	if frameIdx == p.panicAt {
		panic("widget exploded")
	}
}

func TestRootPaintPropagatesWorkerPanic(t *testing.T) {
	for _, parallelism := range []int{1, 4} {
		t.Run("parallelism", func(t *testing.T) {
			root := Root{Child: panickingWidget{frames: 20, panicAt: 7}}

			var recovered any
			func() {
				defer func() { recovered = recover() }()
				for range root.Paint(context.Background(), 64, 32, true, WithMaxParallelFrames(parallelism)) {
				}
			}()

			require.NotNil(t, recovered, "panic in a frame worker must surface on the calling goroutine")

			var perr *PanicError
			require.ErrorAs(t, recovered.(error), &perr)
			assert.Equal(t, "widget exploded", perr.Value)
			assert.Contains(t, string(perr.Stack), "panickingWidget")
		})
	}
}

func TestRootPaintWithoutPanic(t *testing.T) {
	root := Root{Child: panickingWidget{frames: 5, panicAt: -1}}

	count := 0
	require.NotPanics(t, func() {
		for range root.Paint(context.Background(), 64, 32, true, WithMaxParallelFrames(2)) {
			count++
		}
	})
	assert.Equal(t, 5, count)
}
