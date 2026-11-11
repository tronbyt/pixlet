package loader

import (
	"context"
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

// A loader reports each render's sound_id with that render's image, so a
// later render can't overwrite it before the update is read.
func TestLoaderRenderReportsSoundID(t *testing.T) {
	dir := t.TempDir()
	const app = `
load("render.star", "render")
def main(config):
    return render.Root(sound_id = config.get("event", ""), child = render.Box())
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.star"), []byte(app), 0o600))
	root, err := os.OpenRoot(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = root.Close() })

	l, err := NewLoader("app", root, false, nil, nil, "")
	require.NoError(t, err)

	img, first, err := l.renderApplet(t.Context(), map[string]any{"event": "goal-1-0"})
	require.NoError(t, err)
	assert.NotEmpty(t, img)
	_, second, err := l.renderApplet(t.Context(), map[string]any{"event": "goal-2-0"})
	require.NoError(t, err)

	assert.Equal(t, "goal-1-0", first.SoundID)
	assert.Equal(t, "goal-2-0", second.SoundID)
}

// The on-demand path reports the sound_id of the render it returns.
func TestLoadAppletWithResultReportsSoundID(t *testing.T) {
	dir := t.TempDir()
	const app = `
load("render.star", "render")
def main(config):
    return render.Root(sound_id = config.get("event", ""), child = render.Box())
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.star"), []byte(app), 0o600))
	root, err := os.OpenRoot(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = root.Close() })

	l, err := NewLoader("app", root, false, nil, nil, "")
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	go func() { _ = l.Run(ctx) }()

	img, result, err := l.LoadAppletWithResult(map[string]any{"event": "goal-3-1"})
	require.NoError(t, err)
	assert.NotEmpty(t, img)
	assert.Equal(t, "goal-3-1", result.SoundID)

	_, result, err = l.LoadAppletWithResult(map[string]any{})
	require.NoError(t, err)
	assert.Empty(t, result.SoundID)
}
