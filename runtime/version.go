package runtime

import (
	"runtime/debug"
	"strings"
)

var Version = "dev"

func init() { //nolint:gochecknoinits
	if Version != "dev" {
		return
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	var v string
	if strings.HasSuffix(info.Main.Path, "github.com/tronbyt/pixlet") {
		v = info.Main.Version
	} else {
		for _, dep := range info.Deps {
			if strings.HasSuffix(dep.Path, "github.com/tronbyt/pixlet") {
				v = dep.Version
				break
			}
		}
	}

	if v != "" && v != "(devel)" {
		Version = v
	}
}
