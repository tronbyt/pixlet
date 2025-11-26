package canvas

import (
	"errors"

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
}

func (c Metadata) ScaledWidth() int {
	if c.Is2x {
		return c.Width * 2
	}
	return c.Width
}

func (c Metadata) ScaledHeight() int {
	if c.Is2x {
		return c.Height * 2
	}
	return c.Height
}
