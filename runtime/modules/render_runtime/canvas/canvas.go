package canvas

import (
	"errors"
	"time"

	"go.starlark.net/starlark"
)

const threadCanvasKey = "github.com/tronbyt/pixlet/runtime/$canvas"

func AttachToThread(t *starlark.Thread, m Metadata) {
	t.SetLocal(threadCanvasKey, m)
}

var ErrNoCanvas = errors.New("no canvas metadata available")

func FromThread(thread *starlark.Thread) (Metadata, error) {
	if thread == nil {
		return Metadata{}, ErrNoCanvas
	}
	m, ok := thread.Local(threadCanvasKey).(Metadata)
	if !ok {
		return Metadata{}, ErrNoCanvas
	}
	return m, nil
}

type Metadata struct {
	Width  int  `json:"width"`
	Height int  `json:"height"`
	Is2x   bool `json:"is2x"`

	// MaxDuration is the ceiling the encoder puts on the finished animation:
	// it stops emitting frames once their cumulative duration reaches this, so
	// anything past it is never written to the file at all. Apps that generate
	// their own timeline (cycling cards, a long Marquee) need it to size that
	// timeline to fit instead of having the tail silently dropped.
	//
	// Zero means no ceiling. Callers that set a real limit should set it here
	// too; NewRenderConfig does that automatically from WithMaxDuration.
	MaxDuration time.Duration `json:"max_duration"`
}

func (c Metadata) ScaledWidth() int {
	if c.Is2x {
		return c.Width * 2
	}
	return c.Width
}

// MaxDurationMillis returns MaxDuration in whole milliseconds -- the unit
// render.Root's delay is expressed in, and so the one apps do arithmetic in.
// Zero means unbounded, not "no time available".
func (c Metadata) MaxDurationMillis() int {
	return int(c.MaxDuration / time.Millisecond)
}

func (c Metadata) ScaledHeight() int {
	if c.Is2x {
		return c.Height * 2
	}
	return c.Height
}
