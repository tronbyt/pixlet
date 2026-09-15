package runtime

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
)

// An app that lays out its own timeline needs the encoder's ceiling to size it,
// so canvas.max_duration_ms() has to reach Starlark from the thread metadata.
func TestCanvasMaxDurationMillis(t *testing.T) {
	for _, tc := range []struct {
		name string
		meta canvas.Metadata
		want int
	}{
		{"ceiling is reported in ms", canvas.Metadata{Width: 64, Height: 32, MaxDuration: 20 * time.Second}, 20000},
		{"no ceiling reports zero", canvas.Metadata{Width: 64, Height: 32}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := `
load("render.star", "render", "canvas")
load("assert.star", "assert")

assert.eq(canvas.max_duration_ms(), ` + strconv.Itoa(tc.want) + `)

def main():
    return render.Root(child = render.Box())
`
			app, err := NewApplet(t.Context(), "max_duration.star", []byte(src),
				WithTests(t), WithCanvasMeta(tc.meta))
			require.NoError(t, err)
			screens, err := app.Run(t.Context())
			require.NoError(t, err)
			assert.NotNil(t, screens)
		})
	}
}
