package loader

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderSource(t *testing.T, src string, options ...Option) (RenderResult, error) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.star"), []byte(src), 0o600))
	return RenderApplet(t.Context(), filepath.Join(dir, "app.star"), nil, options...)
}

func TestRenderAppletReportsSoundID(t *testing.T) {
	const app = `
load("render.star", "render")
def main():
    return render.Root(sound_id = "goal-2-1", child = render.Box())
`
	result, err := renderSource(t, app)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Image)
	assert.Equal(t, "goal-2-1", result.SoundID)
}

func TestRenderAppletReportsNoSoundID(t *testing.T) {
	const app = `
load("render.star", "render")
def main():
    return render.Root(child = render.Box())
`
	result, err := renderSource(t, app)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Image)
	assert.Empty(t, result.SoundID)
}

func TestRenderAppletReportsOutput(t *testing.T) {
	const app = `
load("render.star", "render")
def main():
    print("hello")
    return render.Root(child = render.Box())
`
	result, err := renderSource(t, app, WithSilenceOutput(true))
	require.NoError(t, err)
	assert.Equal(t, []string{"hello"}, result.Output)
}

func TestRenderAppletKeepsOutputOnError(t *testing.T) {
	const app = `
def main():
    print("before failing")
    fail("broken")
`
	result, err := renderSource(t, app, WithSilenceOutput(true))
	require.Error(t, err)
	assert.Empty(t, result.Image)
	assert.Equal(t, []string{"before failing"}, result.Output)
}

// Output printed by top-level code is kept when the app fails to load.
func TestRenderAppletKeepsOutputOnLoadError(t *testing.T) {
	const app = `
print("loading")
fail("broken at load")

def main():
    return None
`
	result, err := renderSource(t, app, WithSilenceOutput(true))
	require.Error(t, err)
	assert.Empty(t, result.Image)
	assert.Equal(t, []string{"loading"}, result.Output)
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

	first, err := l.renderApplet(t.Context(), map[string]any{"event": "goal-1-0"})
	require.NoError(t, err)
	assert.NotEmpty(t, first.Image)
	second, err := l.renderApplet(t.Context(), map[string]any{"event": "goal-2-0"})
	require.NoError(t, err)

	assert.Equal(t, "goal-1-0", first.SoundID)
	assert.Equal(t, "goal-2-0", second.SoundID)
}

// LoadApplet renders with the config it was given, even when the loader picks
// up the render request before the config.
func TestLoadAppletUsesItsConfig(t *testing.T) {
	dir := t.TempDir()
	const app = `
load("render.star", "render")
def main(config):
    if not config.get("event"):
        fail("missing config")
    return render.Root(child = render.Box())
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

	for range 20 {
		img, err := l.LoadApplet(map[string]any{"event": "goal-3-1"})
		require.NoError(t, err)
		assert.NotEmpty(t, img)
	}
}
