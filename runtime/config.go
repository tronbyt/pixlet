package runtime

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/tronbyt/pixlet/schema"
	"go.starlark.net/starlark"
)

func NewAppletConfig(config map[string]any, sch *schema.Schema) AppletConfig {
	var defaults map[string]string
	if sch != nil {
		defaults = make(map[string]string, len(sch.Fields))
		for _, field := range sch.Fields {
			defaults[field.ID] = field.Default
		}
	}

	return AppletConfig{config, defaults}
}

type AppletConfig struct {
	config   map[string]any
	defaults map[string]string
}

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

func (a AppletConfig) get(key string) (value any, isDefault bool, ok bool) {
	if val, found := a.config[key]; found {
		return val, false, found
	}

	if a.defaults != nil {
		if val, found := a.defaults[key]; found && val != "" {
			return val, true, true
		}
	}

	return nil, false, false
}

func (a AppletConfig) Get(key starlark.Value) (starlark.Value, bool, error) {
	strKey, ok := key.(starlark.String)
	if !ok {
		return nil, false, nil
	}

	val, _, found := a.get(strKey.GoString())
	if !found {
		return nil, false, nil
	}

	strVal, ok := normalizeValue(val)
	return starlark.String(strVal), ok, nil
}

func (a AppletConfig) String() string       { return "AppletConfig(...)" }
func (a AppletConfig) Type() string         { return "AppletConfig" }
func (a AppletConfig) Freeze()              {}
func (a AppletConfig) Truth() starlark.Bool { return true }

func (a AppletConfig) Hash() (uint32, error) {
	sum, err := hashstructure.Hash(a, hashstructure.FormatV2, nil)
	return uint32(sum), err
}

func (a AppletConfig) getString(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key starlark.String
	var def starlark.Value = starlark.None

	if err := starlark.UnpackPositionalArgs(
		"str", args, kwargs, 1,
		&key, &def,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for config.str: %v", err)
	}

	if val, isDefault, ok := a.get(key.GoString()); ok && (!isDefault || len(args) == 1) {
		if str, ok := normalizeValue(val); ok {
			return starlark.String(str), nil
		}
	}

	return def, nil
}

func (a AppletConfig) getBoolean(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key starlark.String
	var def starlark.Value = starlark.None

	if err := starlark.UnpackPositionalArgs(
		"bool", args, kwargs, 1,
		&key, &def,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for config.bool: %v", err)
	}

	if val, isDefault, ok := a.get(key.GoString()); ok && (!isDefault || len(args) == 1) {
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
