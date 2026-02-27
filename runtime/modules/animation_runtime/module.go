package animation_runtime

import (
	"sync"

	"github.com/tronbyt/pixlet/render/animation"
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
			"animation": &starlarkstruct.Module{
				Name:    "render",
				Members: newAnimations(),
			},
		}
	})

	return module.module, nil
}

type transformUnwrapper interface {
	AsAnimationTransform() animation.Transform
}
