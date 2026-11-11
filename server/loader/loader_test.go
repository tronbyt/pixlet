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
