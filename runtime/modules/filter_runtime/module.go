package filter_runtime

import (
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type Module struct {
	once   sync.Once
	module starlark.StringDict
}

var module = Module{}

func LoadModule() (starlark.StringDict, error) {
	module.once.Do(func() {
		module.module = starlark.StringDict{
			"filter": &starlarkstruct.Module{
				Name:    "filter",
				Members: newFilters(),
			},
		}
	})

	return module.module, nil
}
