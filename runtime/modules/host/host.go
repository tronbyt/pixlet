// Package host provides the server.star and device.star modules, which tell an
// app about the server rendering it and the device it is rendering for.
//
// The values come from whoever runs the app (Tronbyt Server, for example)
// through runtime.WithHostInfo, and are fixed when the module is loaded, so
// they also work in callbacks such as custom animation curves. Anything the
// host doesn't provide reads as None, so apps must handle that case, for
// example when run with `pixlet render`.
package host

import (
	"time"

	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ServerModuleName = "server"
	DeviceModuleName = "device"
)

// Info describes the host running an app. Empty fields mean "unknown".
type Info struct {
	// ServerVersion is the host server's version, e.g. "v2.6.0".
	ServerVersion string
	// SystemAppsCommit is the commit of the server's local app store
	// (system apps repository).
	SystemAppsCommit string
	// SystemAppsCommitTime is the time of SystemAppsCommit.
	SystemAppsCommitTime time.Time

	Device Device
}

// Device describes the device an app is rendering for.
type Device struct {
	// Model is the device type, e.g. "tidbyt_gen2".
	Model string
	// FirmwareVersion is the version the device last reported, e.g. "v1.7.4".
	FirmwareVersion string
	// FirmwareType is the kind of firmware, e.g. "esp32" or "tronberry".
	FirmwareType string
}

// LoadServerModule returns the server.star module for info. pixletVersion is
// the version of the Pixlet runtime itself, which is always known.
func LoadServerModule(pixletVersion string, info Info) (starlark.StringDict, error) {
	commitTime := starlark.Value(starlark.None)
	if !info.SystemAppsCommitTime.IsZero() {
		commitTime = starlarktime.Time(info.SystemAppsCommitTime)
	}

	return starlark.StringDict{
		ServerModuleName: &starlarkstruct.Module{
			Name: ServerModuleName,
			Members: starlark.StringDict{
				"version":                 noArgs("version", optString(info.ServerVersion)),
				"pixlet_version":          noArgs("pixlet_version", optString(pixletVersion)),
				"system_apps_commit":      noArgs("system_apps_commit", optString(info.SystemAppsCommit)),
				"system_apps_commit_time": noArgs("system_apps_commit_time", commitTime),
			},
		},
	}, nil
}

// LoadDeviceModule returns the device.star module for info.
func LoadDeviceModule(info Info) (starlark.StringDict, error) {
	return starlark.StringDict{
		DeviceModuleName: &starlarkstruct.Module{
			Name: DeviceModuleName,
			Members: starlark.StringDict{
				"model":            noArgs("model", optString(info.Device.Model)),
				"firmware_version": noArgs("firmware_version", optString(info.Device.FirmwareVersion)),
				"firmware_type":    noArgs("firmware_type", optString(info.Device.FirmwareType)),
			},
		},
	}, nil
}

// noArgs returns a builtin that takes no arguments and returns v.
func noArgs(name string, v starlark.Value) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs(b.Name(), args, kwargs); err != nil {
			return nil, err
		}
		return v, nil
	})
}

func optString(s string) starlark.Value {
	if s == "" {
		return starlark.None
	}
	return starlark.String(s)
}
