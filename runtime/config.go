package runtime

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	"go.starlark.net/starlark"
)

type AppletConfig map[string]any

func (a AppletConfig) AttrNames() []string {
	return []string{
		"get",
		"str",
		"bool",
	}
}

func (a AppletConfig) Attr(name string) (starlark.Value, error) {
	switch name {

	case "get", "str":
		return starlark.NewBuiltin("str", a.getString), nil

	case "bool":
		return starlark.NewBuiltin("bool", a.getBoolean), nil

	default:
		return nil, nil
	}
}

func (a AppletConfig) Get(key starlark.Value) (starlark.Value, bool, error) {
	if v, ok := key.(starlark.String); ok {
		if val, found := a[v.GoString()]; found {
			str, ok := normalizeValue(val)
			return starlark.String(str), ok, nil
		}
	}
	return nil, false, nil
}

func (a AppletConfig) String() string       { return "AppletConfig(...)" }
func (a AppletConfig) Type() string         { return "AppletConfig" }
func (a AppletConfig) Freeze()              {}
func (a AppletConfig) Truth() starlark.Bool { return true }

func (a AppletConfig) Hash() (uint32, error) {
	sum, err := hashstructure.Hash(a, hashstructure.FormatV2, nil)
	return uint32(sum), err
}

func (a AppletConfig) getString(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key starlark.String
	var def starlark.Value
	def = starlark.None

	if err := starlark.UnpackPositionalArgs(
		"str", args, kwargs, 1,
		&key, &def,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for config.str: %v", err)
	}

	if val, ok := a[key.GoString()]; ok {
		if str, ok := normalizeValue(val); ok {
			return starlark.String(str), nil
		}
	}
	return def, nil
}

func (a AppletConfig) getBoolean(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key starlark.String
	var def starlark.Value
	def = starlark.None

	if err := starlark.UnpackPositionalArgs(
		"bool", args, kwargs, 1,
		&key, &def,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for config.bool: %v", err)
	}

	if val, ok := a[key.GoString()]; ok {
		switch val := val.(type) {
		case bool:
			return starlark.Bool(val), nil
		case string:
			b, err := strconv.ParseBool(val)
			return starlark.Bool(b), err
		default:
			if str, ok := normalizeValue(val); ok {
				b, err := strconv.ParseBool(str)
				return starlark.Bool(b), err
			}
		}
	}
	return def, nil
}

func normalizeValue(val any) (string, bool) {
	if val == nil {
		return "", false
	}

	if str, ok := val.(string); ok {
		return str, true
	}

	if b, err := json.Marshal(val); err == nil {
		return string(b), true
	}

	return fmt.Sprintf("%v", val), true
}
