package main

import (
	"image/color"
	"reflect"

	"github.com/tronbyt/pixlet/render"
	"github.com/tronbyt/pixlet/render/animation"
	"github.com/tronbyt/pixlet/render/filter"
)

// Packages is a list of packages and their types to generate code and documentation for.
var Packages = []Package{
	{
		Name:           "render",
		Directory:      "render",
		ImportPath:     "github.com/tronbyt/pixlet/render",
		HeaderTemplate: "header/render.tmpl",
		TypeTemplate:   "type.tmpl",
		CodePath:       "runtime/modules/render_runtime/generated.go",
		DocTemplate:    "docs/render.tmpl",
		DocPath:        "docs/widgets.md",
		GoRootName:     "Root",
		GoWidgetName:   "Widget",
		Types: []reflect.Value{
			reflect.ValueOf(new(render.Animation)),
			reflect.ValueOf(new(render.Arc)),
			reflect.ValueOf(new(render.Box)),
			reflect.ValueOf(new(render.Circle)),
			reflect.ValueOf(new(render.Column)),
			reflect.ValueOf(new(render.Emoji)),
			reflect.ValueOf(new(render.Image)),
			reflect.ValueOf(new(render.Line)),
			reflect.ValueOf(new(render.Marquee)),
			reflect.ValueOf(new(render.Padding)),
			reflect.ValueOf(new(render.PieChart)),
			reflect.ValueOf(new(render.Plot)),
			reflect.ValueOf(new(render.Polygon)),
			reflect.ValueOf(new(render.Root)),
			reflect.ValueOf(new(render.Row)),
			reflect.ValueOf(new(render.Sequence)),
			reflect.ValueOf(new(render.Stack)),
			reflect.ValueOf(new(render.Text)),
			reflect.ValueOf(new(render.WrappedText)),
		},
	},
	{
		Name:           "animation",
		Directory:      "render/animation",
		ImportPath:     "github.com/tronbyt/pixlet/render/animation",
		HeaderTemplate: "header/animation.tmpl",
		TypeTemplate:   "type.tmpl",
		CodePath:       "runtime/modules/animation_runtime/generated.go",
		DocTemplate:    "docs/animation.tmpl",
		DocPath:        "docs/animation.md",
		GoRootName:     "render_runtime.Root",
		GoWidgetName:   "render_runtime.Widget",
		Types: []reflect.Value{
			reflect.ValueOf(new(animation.Keyframe)),
			reflect.ValueOf(new(animation.Origin)),
			reflect.ValueOf(new(animation.Rotate)),
			reflect.ValueOf(new(animation.Scale)),
			reflect.ValueOf(new(animation.Shear)),
			reflect.ValueOf(new(animation.Transformation)),
			reflect.ValueOf(new(animation.Translate)),

			// Legacy
			reflect.ValueOf(new(animation.AnimatedPositioned)),
		},
	},
	{
		Name:           "filter",
		Directory:      "render/filter",
		ImportPath:     "github.com/tronbyt/pixlet/render/filter",
		HeaderTemplate: "header/filters.tmpl",
		TypeTemplate:   "type.tmpl",
		CodePath:       "runtime/modules/filter_runtime/generated.go",
		DocTemplate:    "docs/filters.tmpl",
		DocPath:        "docs/filters.md",
		GoRootName:     "render_runtime.Root",
		GoWidgetName:   "render_runtime.Widget",
		Types: []reflect.Value{
			reflect.ValueOf(new(filter.Blur)),
			reflect.ValueOf(new(filter.Brightness)),
			reflect.ValueOf(new(filter.Contrast)),
			reflect.ValueOf(new(filter.EdgeDetection)),
			reflect.ValueOf(new(filter.Emboss)),
			reflect.ValueOf(new(filter.FlipHorizontal)),
			reflect.ValueOf(new(filter.FlipVertical)),
			reflect.ValueOf(new(filter.Gamma)),
			reflect.ValueOf(new(filter.Grayscale)),
			reflect.ValueOf(new(filter.Hue)),
			reflect.ValueOf(new(filter.Invert)),
			reflect.ValueOf(new(filter.Rotate)),
			reflect.ValueOf(new(filter.Saturation)),
			reflect.ValueOf(new(filter.Sepia)),
			reflect.ValueOf(new(filter.Sharpen)),
			reflect.ValueOf(new(filter.Shear)),
			reflect.ValueOf(new(filter.Threshold)),
		},
	},
}

// TypeMap is a map of Go types to an `Attribute` definition.
var TypeMap = map[reflect.Type]Type{
	// Primitive types
	toDecayedType(new(string)): {
		GoType:       "starlark.String",
		DocType:      "str",
		TemplatePath: "attr/string.tmpl",
	},
	toDecayedType(new(int)): {
		GoType:       "starlark.Int",
		DocType:      "int",
		TemplatePath: "attr/int.tmpl",
	},
	toDecayedType(new(int32)): {
		GoType:       "starlark.Int",
		DocType:      "int",
		TemplatePath: "attr/int32.tmpl",
	},
	toDecayedType(new(float64)): {
		GoType:       "starlark.Value",
		DocType:      "float / int",
		TemplatePath: "attr/float.tmpl",
	},
	toDecayedType(new(bool)): {
		GoType:       "starlark.Bool",
		DocType:      "bool",
		TemplatePath: "attr/bool.tmpl",
	},

	// Render types
	toDecayedType(new(render.Insets)): {
		GoType:       "starlark.Value",
		DocType:      "int / tuple of 3 ints",
		TemplatePath: "attr/insets.tmpl",
	},
	toDecayedType(new(render.Widget)): {
		GoType:       "starlark.Value",
		DocType:      "Widget",
		TemplatePath: "attr/child.tmpl",
	},
	toDecayedType(new([]render.Widget)): {
		GoType:       "*starlark.List",
		DocType:      "[Widget]",
		TemplatePath: "attr/children.tmpl",
	},
	toDecayedType(new(color.Color)): {
		GoType:        "starlark.Value",
		DocType:       `color`,
		TemplatePath:  "attr/color.tmpl",
		GenerateField: true,
		DefaultValue:  "starlark.None",
	},

	// Render `PieChart types`
	toDecayedType(new([]color.Color)): {
		GoType:        "*starlark.List",
		DocType:       `[color]`,
		TemplatePath:  "attr/colors.tmpl",
		GenerateField: true,
	},
	toDecayedType(new([]float64)): {
		GoType:        "*starlark.List",
		DocType:       `[float]`,
		TemplatePath:  "attr/weights.tmpl",
		GenerateField: true,
	},

	// Render `Plot` types`
	toDecayedType(new([2]float64)): {
		GoType:       "starlark.Tuple",
		DocType:      "(float, float)",
		TemplatePath: "attr/datapoint.tmpl",
	},
	toDecayedType(new([][2]float64)): {
		GoType:       "*starlark.List",
		DocType:      "[(float, float)]",
		TemplatePath: "attr/dataseries.tmpl",
	},

	// Animation types
	toDecayedType(new(animation.Origin)): {
		GoType:       "starlark.Value",
		DocType:      "Origin",
		TemplatePath: "attr/origin.tmpl",
	},
	toDecayedType(new(animation.Curve)): {
		GoType:       "starlark.Value",
		DocType:      `str / function`,
		TemplatePath: "attr/curve.tmpl",
	},
	toDecayedType(new(animation.Direction)): {
		GoType:        "starlark.String",
		DocType:       `str`,
		TemplatePath:  "attr/direction.tmpl",
		GenerateField: true,
	},
	toDecayedType(new(animation.FillMode)): {
		GoType:        "starlark.String",
		DocType:       `str`,
		TemplatePath:  "attr/fill_mode.tmpl",
		GenerateField: true,
	},
	toDecayedType(new(animation.Rounding)): {
		GoType:        "starlark.String",
		DocType:       `str`,
		TemplatePath:  "attr/rounding.tmpl",
		GenerateField: true,
	},
	toDecayedType(new(animation.Percentage)): {
		GoType:       "starlark.Value",
		DocType:      `float`,
		TemplatePath: "attr/percentage.tmpl",
	},
	toDecayedType(new([]animation.Keyframe)): {
		GoType:       "*starlark.List",
		DocType:      "[Keyframe]",
		TemplatePath: "attr/keyframes.tmpl",
	},
	toDecayedType(new([]animation.Transform)): {
		GoType:       "*starlark.List",
		DocType:      "[Transform]",
		TemplatePath: "attr/transforms.tmpl",
	},
	toDecayedType(new([]render.Point)): {
		GoType:        "*starlark.List",
		DocType:       `[(float, float)]`,
		TemplatePath:  "attr/vertices.tmpl",
		GenerateField: true,
	},
}
