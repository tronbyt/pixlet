package animation_runtime

import (
	"sync"

	"github.com/tronbyt/pixlet/render/animation"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type AnimationModule struct {
	once   sync.Once
	module starlark.StringDict
}

var animationModule = AnimationModule{}

func LoadAnimationModule() (starlark.StringDict, error) {
	animationModule.once.Do(func() {
		animationModule.module = starlark.StringDict{
			"animation": &starlarkstruct.Module{
				Name:    "render",
				Members: newAnimations(),
			},
		}
	})

	return animationModule.module, nil
}

type transformUnwrapper interface {
	AsAnimationTransform() animation.Transform
}
