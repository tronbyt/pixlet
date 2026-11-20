package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/runtime/modules/host"
)

func TestRenderAppletPassesHostInfo(t *testing.T) {
	dir := t.TempDir()
	const app = `
load("device.star", "device")
load("render.star", "render")
load("server.star", "server")
def main():
    if server.version() != "v2.6.0" or device.firmware_version() != "v1.7.4":
        fail("host info not passed through")
    return render.Root(child = render.Box())
`
	path := filepath.Join(dir, "app.star")
	require.NoError(t, os.WriteFile(path, []byte(app), 0o600))
	img, _, err := RenderApplet(t.Context(), path, nil, WithHostInfo(host.Info{
		ServerVersion: "v2.6.0",
		Device:        host.Device{FirmwareVersion: "v1.7.4"},
	}))
	require.NoError(t, err)
	require.NotEmpty(t, img)
}

func TestNewLoaderPassesHostInfo(t *testing.T) {
	dir := t.TempDir()
	const app = `
load("device.star", "device")
load("render.star", "render")
def main():
    if device.model() != "tidbyt_gen2":
        fail("host info not passed through")
    return render.Root(child = render.Box())
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.star"), []byte(app), 0o600))
	root, err := os.OpenRoot(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = root.Close() })

	l, err := NewLoader("app", root, false, nil, nil, "", WithHostInfo(host.Info{
		Device: host.Device{Model: "tidbyt_gen2"},
	}))
	require.NoError(t, err)
	_, err = l.applet.Run(t.Context())
	require.NoError(t, err)
}
