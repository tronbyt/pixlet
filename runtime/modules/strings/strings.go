package strings

import (
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const ModuleName = "strings"

var (
	once   sync.Once
	module starlark.StringDict
)

func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		module = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"pad":      starlark.NewBuiltin("pad", pad),
					"truncate": starlark.NewBuiltin("truncate", truncate),
				},
			},
		}
		module.Freeze()
	})
	return module, nil
}
