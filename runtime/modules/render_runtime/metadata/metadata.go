package metadata

import (
	"errors"

	"go.starlark.net/starlark"
)

const threadMetadataKey = "github.com/tronbyt/pixlet/runtime/$metadata"

func AttachToThread(t *starlark.Thread, m Metadata) {
	t.SetLocal(threadMetadataKey, m)
}

var ErrNoMetadata = errors.New("no metadata available")

func FromThread(thread *starlark.Thread) (Metadata, error) {
	if thread == nil {
		return Metadata{}, ErrNoMetadata
	}
	m, ok := thread.Local(threadMetadataKey).(Metadata)
	if !ok {
		return Metadata{}, ErrNoMetadata
	}
	return m, nil
}

type Metadata struct {
	Width  int
	Height int
	Is2x   bool
}

func (m Metadata) ScaledWidth() int {
	if m.Is2x {
		return m.Width * 2
	}
	return m.Width
}

func (m Metadata) ScaledHeight() int {
	if m.Is2x {
		return m.Height * 2
	}
	return m.Height
}
