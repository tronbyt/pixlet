package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderSource(t *testing.T, src string, options ...Option) ([]byte, error) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.star"), []byte(src), 0o600))
	img, _, err := RenderApplet(t.Context(), filepath.Join(dir, "app.star"), nil, options...)
	return img, err
}

func TestRenderAppletReportsSoundID(t *testing.T) {
	const app = `
load("render.star", "render")
def main():
    return render.Root(sound_id = "goal-2-1", child = render.Box())
`
	var result RenderResult
	img, err := renderSource(t, app, WithResult(&result))
	require.NoError(t, err)
	assert.NotEmpty(t, img)
	assert.Equal(t, "goal-2-1", result.SoundID)
}

func TestRenderAppletReportsNoSoundID(t *testing.T) {
	const app = `
load("render.star", "render")
def main():
    return render.Root(child = render.Box())
`
	result := RenderResult{SoundID: "stale"}
	_, err := renderSource(t, app, WithResult(&result))
	require.NoError(t, err)
	assert.Empty(t, result.SoundID)
}

func TestRenderAppletWithoutResultOption(t *testing.T) {
	const app = `
load("render.star", "render")
def main():
    return render.Root(sound_id = "goal-2-1", child = render.Box())
`
	img, err := renderSource(t, app)
	require.NoError(t, err)
	assert.NotEmpty(t, img)
}
