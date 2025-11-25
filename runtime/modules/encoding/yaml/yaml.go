package yaml

import (
	"fmt"
	"strings"
	"sync"

	"github.com/qri-io/starlib/util"
	"github.com/tronbyt/pixlet/starlarkutil"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"gopkg.in/yaml.v3"
)

const (
	ModuleName    = "encoding/yaml.star"
	DefaultIndent = 2
)

var (
	once       sync.Once
	yamlModule starlark.StringDict
)

// LoadModule loads the yaml module.
// It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		yamlModule = starlark.StringDict{
			"yaml": &starlarkstruct.Module{
				Name: "yaml",
				Members: starlark.StringDict{
					"decode": starlark.NewBuiltin("decode", Decode),
					"encode": starlark.NewBuiltin("encode", Encode),
				},
			},
		}
	})
	return yamlModule, nil
}

// Decode reads the YAML-encoded value from its input and returns the result as a starlark.Value.
func Decode(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.String

	err := starlark.UnpackArgs(
		"decode", args, kwargs,
		"x", &x,
	)
	if err != nil {
		return nil, err
	}

	r := strings.NewReader(x.GoString())

	var val any
	if err := yaml.NewDecoder(r).Decode(&val); err != nil {
		return starlark.None, err
	}

	return util.Marshal(val)
}

// Encode returns the YAML encoding of its input.
func Encode(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		x      starlark.Value
		indent starlark.Int
	)

	err := starlark.UnpackArgs(
		"encode", args, kwargs,
		"x", &x,
		"indent?", &indent,
	)
	if err != nil {
		return starlark.None, err
	}

	val, err := util.Unmarshal(x)
	if err != nil {
		return starlark.None, err
	}

	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	defer enc.Close()

	indentVal, err := starlarkutil.AsInt64(indent)
	if err != nil {
		return starlark.None, fmt.Errorf("parsing indent: %w", err)
	}

	if indentVal == 0 {
		indentVal = DefaultIndent
	}

	enc.SetIndent(int(indentVal))

	if err := enc.Encode(val); err != nil {
		return starlark.None, err
	}

	return starlark.String(buf.String()), nil
}
