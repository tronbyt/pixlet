package host_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/runtime/modules/host"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func run(t *testing.T, src string, opts ...runtime.AppletOption) {
	t.Helper()
	opts = append(opts, runtime.WithTests(t))
	app, err := runtime.NewApplet(t.Context(), "test.star", []byte(src), opts...)
	require.NoError(t, err)
	_, err = app.Run(t.Context())
	require.NoError(t, err)
}

func TestHostInfo(t *testing.T) {
	updated := time.Date(2026, 11, 10, 4, 38, 0, 0, time.UTC)
	run(t, `
load("assert.star", "assert")
load("device.star", "device")
load("render.star", "render")
load("server.star", "server")

def main():
    assert.eq(server.version(), "v2.6.0")
    assert.true(server.pixlet_version())
    assert.eq(server.system_apps_commit(), "5fb4e4ba6")
    assert.eq(server.system_apps_commit_time().year, 2026)

    assert.eq(device.model(), "tidbyt_gen2")
    assert.eq(device.firmware_version(), "v1.7.4")
    assert.eq(device.firmware_type(), "esp32")
    return render.Root(child = render.Box())
`, runtime.WithHostInfo(host.Info{
		ServerVersion:        "v2.6.0",
		SystemAppsCommit:     "5fb4e4ba6",
		SystemAppsCommitTime: updated,
		Device: host.Device{
			Model:           "tidbyt_gen2",
			FirmwareVersion: "v1.7.4",
			FirmwareType:    "esp32",
		},
	}))
}

func TestHostInfoUnknown(t *testing.T) {
	run(t, `
load("assert.star", "assert")
load("device.star", "device")
load("render.star", "render")
load("server.star", "server")

def main():
    assert.eq(server.version(), None)
    assert.eq(server.system_apps_commit(), None)
    assert.eq(server.system_apps_commit_time(), None)
    assert.eq(device.model(), None)
    assert.eq(device.firmware_version(), None)
    assert.eq(device.firmware_type(), None)
    return render.Root(child = render.Box())
`)
}

func TestHostInfoRejectsArguments(t *testing.T) {
	app, err := runtime.NewApplet(t.Context(), "test.star", []byte(`
load("server.star", "server")
def main():
    return server.version("x")
`), runtime.WithTests(t))
	require.NoError(t, err)
	_, err = app.Run(t.Context())
	require.Error(t, err)
}

// Callbacks such as custom animation curves run on bare threads, so the
// values must not depend on the calling thread.
func TestHostInfoOnBareThread(t *testing.T) {
	mod, err := host.LoadDeviceModule(host.Info{Device: host.Device{Model: "tidbyt_gen2"}})
	require.NoError(t, err)
	model, err := mod[host.DeviceModuleName].(*starlarkstruct.Module).Attr("model")
	require.NoError(t, err)
	v, err := starlark.Call(&starlark.Thread{}, model, nil, nil)
	require.NoError(t, err)
	require.Equal(t, starlark.String("tidbyt_gen2"), v)
}
