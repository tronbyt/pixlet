package time_runtime

import (
	"maps"
	"sync"
	"time"

	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName        = "time.star"
	threadTimezoneKey = "github.com/tronbyt/pixlet/runtime/$timezone"
)

func SetLocation(t *starlark.Thread, tz *time.Location) {
	if tz != nil {
		t.SetLocal(threadTimezoneKey, tz)
		starlarktime.SetNow(t, func() (time.Time, error) {
			return time.Now().In(tz), nil
		})
	}
}

func GetLocation(t *starlark.Thread) *time.Location {
	if t != nil {
		if v, ok := t.Local(threadTimezoneKey).(*time.Location); ok {
			return v
		}
	}
	return time.Local
}

var (
	once   sync.Once
	module starlark.StringDict
)

func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		members := maps.Clone(starlarktime.Module.Members)
		members["tz"] = starlark.NewBuiltin("tz", tz)

		module = starlark.StringDict{
			"time": &starlarkstruct.Module{
				Name:    starlarktime.Module.Name,
				Members: members,
			},
		}
	})
	return module, nil
}

func tz(thread *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	return starlark.String(GetLocation(thread).String()), nil
}
