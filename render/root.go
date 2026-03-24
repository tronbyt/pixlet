package render

import (
	"context"
	"image"
	"image/color"
	"iter"
	"runtime"
	"sync"

	"github.com/tronbyt/gg"
)

const (
	// DefaultFrameWidth is the normal width for a frame.
	DefaultFrameWidth = 64

	// DefaultFrameHeight is the normal height for a frame.
	DefaultFrameHeight = 32

	// DefaultMaxFrameCount is the default maximum number of frames to render.
	DefaultMaxFrameCount = 2000
)

// Root is the top level of every Widget tree.
//
// The child widget, and all its descendants, will be drawn on a 64x32
// canvas. Root places its child in the upper left corner of the
// canvas.
//
// If the tree contains animated widgets, the resulting animation will
// run with _delay_ milliseconds per frame.
//
// If the tree holds time sensitive information which must never be
// displayed past a certain point in time, pass _MaxAge_ to specify
// an expiration time in seconds. Display devices use this to avoid
// displaying stale data in the event of e.g. connectivity issues.
type Root struct {
	// Widget to render
	Child Widget `starlark:"child,required"`
	// Frame delay in milliseconds
	Delay int32 `starlark:"delay"`
	// Expiration time in seconds
	MaxAge int32 `starlark:"max_age"`
	// Request animation is shown in full, regardless of app cycle speed.
	ShowFullAnimation bool `starlark:"show_full_animation"`

	maxParallelFrames int
	maxFrameCount     int
}

type RootPaintOption func(*Root)

// WithMaxParallelFrames sets the maximum number of frames rendered concurrently.
// If <=0, Paint uses runtime.NumCPU().
func WithMaxParallelFrames(maxFrames int) RootPaintOption {
	return func(r *Root) {
		r.maxParallelFrames = maxFrames
	}
}

// WithMaxFrameCount sets the maximum number of frames that will be rendered.
// If a widget tree has more frames than this, the number of frames will be
// capped.
func WithMaxFrameCount(maxFrames int) RootPaintOption {
	return func(r *Root) {
		r.maxFrameCount = maxFrames
	}
}

// Paint renders the child widget onto the frame. It doesn't do
// any resizing or alignment.
func (r Root) Paint(ctx context.Context, width, height int, solidBackground bool, opts ...RootPaintOption) iter.Seq[image.Image] {
	return func(yield func(image.Image) bool) {
		for _, opt := range opts {
			opt(&r)
		}

		if r.maxFrameCount <= 0 {
			r.maxFrameCount = DefaultMaxFrameCount
		}

		numFrames := min(r.Child.FrameCount(image.Rect(0, 0, width, height)), r.maxFrameCount)
		if numFrames == 0 {
			return
		}

		parallelism := r.maxParallelFrames
		if parallelism <= 0 {
			parallelism = runtime.NumCPU()
		}
		parallelism = max(1, min(parallelism, numFrames))

		var wg sync.WaitGroup
		defer wg.Wait()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		type frameResult struct {
			index int
			img   image.Image
		}

		jobs := make(chan int)
		results := make(chan frameResult, parallelism)
		sema := make(chan struct{}, parallelism)

		// Spawn parallel renderers
		for range parallelism {
			wg.Go(func() {
				for frame := range jobs {
					dc := gg.NewContext(width, height)
					if solidBackground {
						dc.SetColor(color.Black)
						dc.Clear()
					}

					dc.Push()
					r.Child.Paint(dc, image.Rect(0, 0, width, height), frame)
					dc.Pop()

					select {
					case <-ctx.Done():
						return
					case results <- frameResult{index: frame, img: dc.Image()}:
					}
				}
			})
		}

		// Queue a job for each frame
		wg.Go(func() {
			defer close(jobs)
			for i := range numFrames {
				select {
				case <-ctx.Done():
					return
				case sema <- struct{}{}:
				}

				select {
				case <-ctx.Done():
					return
				case jobs <- i:
				}
			}
		})

		// Yield rendered images in order
		pending := make(map[int]image.Image, parallelism)
		for nextToYield := 0; nextToYield < numFrames; {
			var res frameResult
			select {
			case <-ctx.Done():
				return
			case res = <-results:
			}

			if res.index == nextToYield {
				if !yield(res.img) {
					return
				}
				<-sema
				nextToYield++

				for {
					img, ok := pending[nextToYield]
					if !ok {
						break
					}

					delete(pending, nextToYield)
					if !yield(img) {
						return
					}
					<-sema
					nextToYield++
				}
			} else {
				pending[res.index] = res.img
			}
		}
	}
}
