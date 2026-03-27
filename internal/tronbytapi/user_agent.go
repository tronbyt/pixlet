package tronbytapi

import (
	"runtime"
	"runtime/debug"

	pixletruntime "github.com/tronbyt/pixlet/runtime"
)

func commitFromVCS() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) > 7 {
					return setting.Value[:7]
				}
				return setting.Value
			}
		}
	}
	return ""
}

func BuildUserAgent() string {
	version := pixletruntime.Version
	if commit := commitFromVCS(); commit != "" {
		version += "-" + commit
	}
	return "pixlet/" + version + " (" + runtime.GOOS + "/" + runtime.GOARCH + ")"
}
