// Package host provides the server.star and device.star modules, which tell an
// app about the server rendering it and the device it is rendering for.
//
// The values come from whoever runs the app (Tronbyt Server, for example)
// through runtime.WithHostInfo. Anything the host doesn't provide reads as
// None, so apps must handle that case, for example when run with
// `pixlet render`.
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

	threadHostKey = "github.com/tronbyt/pixlet/runtime/$host"
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

// AttachToThread makes info available to the server and device modules.
func AttachToThread(t *starlark.Thread, info Info) {
	t.SetLocal(threadHostKey, info)
}

func fromThread(t *starlark.Thread) Info {
	if t != nil {
		if info, ok := t.Local(threadHostKey).(Info); ok {
			return info
		}
	}
	return Info{}
}

// LoadServerModule returns the server.star module. pixletVersion is the
// version of the Pixlet runtime itself, which is always known.
func LoadServerModule(pixletVersion string) (starlark.StringDict, error) {
	return starlark.StringDict{
		ServerModuleName: &starlarkstruct.Module{
			Name: ServerModuleName,
			Members: starlark.StringDict{
				"version": noArgs("version", func(i Info) starlark.Value { return optString(i.ServerVersion) }),
				"pixlet_version": starlark.NewBuiltin("pixlet_version", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
					if err := starlark.UnpackArgs(b.Name(), args, kwargs); err != nil {
						return nil, err
					}
					return optString(pixletVersion), nil
				}),
				"system_apps_commit": noArgs("system_apps_commit", func(i Info) starlark.Value { return optString(i.SystemAppsCommit) }),
				"system_apps_commit_time": noArgs("system_apps_commit_time", func(i Info) starlark.Value {
					if i.SystemAppsCommitTime.IsZero() {
						return starlark.None
					}
					return starlarktime.Time(i.SystemAppsCommitTime)
				}),
			},
		},
	}, nil
}

// LoadDeviceModule returns the device.star module.
func LoadDeviceModule() (starlark.StringDict, error) {
	return starlark.StringDict{
		DeviceModuleName: &starlarkstruct.Module{
			Name: DeviceModuleName,
			Members: starlark.StringDict{
				"model":            noArgs("model", func(i Info) starlark.Value { return optString(i.Device.Model) }),
				"firmware_version": noArgs("firmware_version", func(i Info) starlark.Value { return optString(i.Device.FirmwareVersion) }),
				"firmware_type":    noArgs("firmware_type", func(i Info) starlark.Value { return optString(i.Device.FirmwareType) }),
			},
		},
	}, nil
}

func noArgs(name string, fn func(Info) starlark.Value) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs(b.Name(), args, kwargs); err != nil {
			return nil, err
		}
		return fn(fromThread(thread)), nil
	})
}

func optString(s string) starlark.Value {
	if s == "" {
		return starlark.None
	}
	return starlark.String(s)
}
