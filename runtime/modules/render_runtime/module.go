package render_runtime

import (
	"sync"

	"github.com/tronbyt/pixlet/render"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type RenderModule struct {
	once   sync.Once
	module starlark.StringDict
}

var renderModule = RenderModule{}

func LoadRenderModule() (starlark.StringDict, error) {
	var err error

	renderModule.once.Do(func() {
		widgets := newWidgets()

		var fontList []string
		if fontList, err = render.GetFontList(); err != nil {
			return
		}
		fnt := starlark.NewDict(len(fontList))
		for _, name := range fontList {
			if err = fnt.SetKey(starlark.String(name), starlark.String(name)); err != nil {
				return
			}
		}
		fnt.Freeze()

		widgets["fonts"] = fnt

		renderModule.module = starlark.StringDict{
			"render": &starlarkstruct.Module{
				Name:    "render",
				Members: widgets,
			},
			"canvas": &starlarkstruct.Module{
				Name: "canvas",
				Members: starlark.StringDict{
					"width":  starlark.NewBuiltin("width", dimension(dimensionWidth)),
					"height": starlark.NewBuiltin("height", dimension(dimensionHeight)),
					"size":   starlark.NewBuiltin("size", size),
					"is2x":   starlark.NewBuiltin("is2x", is2x),
				},
			},
		}
	})

	return renderModule.module, err
}

type Rootable interface {
	AsRenderRoot() render.Root
}

type Widget interface {
	AsRenderWidget() render.Widget
}
