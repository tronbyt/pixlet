package schema

import (
	"fmt"

	"github.com/gohugoio/hashstructure"
	"github.com/tronbyt/pixlet/starlarkutil"
	"go.starlark.net/starlark"
)

type OAuth2 struct {
	SchemaField

	starlarkScopes              *starlark.List
	starlarkAuthorizationParams *starlark.Dict
}

func newOAuth2(
	thread *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var (
		id            starlark.String
		name          starlark.String
		desc          starlark.String
		icon          starlark.String
		handler       *starlark.Function
		clientID      starlark.String
		authEndpoint  starlark.String
		tokenEndpoint starlark.String
		authParams    *starlark.Dict
		scopes        *starlark.List
	)

	if err := starlark.UnpackArgs(
		"OAuth2",
		args, kwargs,
		"id", &id,
		"name", &name,
		"desc", &desc,
		"icon", &icon,
		"handler?", &handler,
		"client_id?", &clientID,
		"authorization_endpoint", &authEndpoint,
		"scopes", &scopes,
		"token_endpoint?", &tokenEndpoint,
		"authorization_params?", &authParams,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for OAuth2: %s", err)
	}

	s := &OAuth2{}
	s.SchemaField.Type = "oauth2"
	s.ID = id.GoString()
	s.Name = name.GoString()
	s.Description = desc.GoString()
	s.Icon = icon.GoString()
	if handler != nil {
		s.Handler = handler.Name()
		s.StarlarkHandler = handler
	}
	s.ClientID = clientID.GoString()
	s.AuthorizationEndpoint = authEndpoint.GoString()
	s.TokenEndpoint = tokenEndpoint.GoString()
	s.starlarkScopes = scopes
	s.starlarkAuthorizationParams = authParams

	if authParams != nil {
		s.AuthorizationParams = make(map[string]string, authParams.Len())
		for _, item := range authParams.Items() {
			key, ok := item[0].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("expected authorization_params keys to be strings but found: %s", item[0].Type())
			}
			value, ok := item[1].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("expected authorization_params values to be strings but found: %s (for %q)", item[1].Type(), key.GoString())
			}
			s.AuthorizationParams[key.GoString()] = value.GoString()
		}
	}

	if s.starlarkScopes != nil {
		s.Scopes = make([]string, 0, scopes.Len())

		for i, scopeVal := range starlarkutil.Enumerate(s.starlarkScopes) {
			if _, isNone := scopeVal.(starlark.NoneType); isNone {
				continue
			}

			scope, ok := scopeVal.(starlark.String)
			if !ok {
				return nil, fmt.Errorf(
					"expected fields to be a list of string but found: %s (at index %d)",
					scopeVal.Type(), i,
				)
			}

			s.Scopes = append(s.Scopes, scope.GoString())
		}
	}

	return s, nil
}

func (s *OAuth2) AsSchemaField() SchemaField {
	return s.SchemaField
}

func (s *OAuth2) AttrNames() []string {
	return []string{
		"id", "name", "desc", "icon", "handler", "client_id", "authorization_endpoint", "scopes",
		"token_endpoint", "authorization_params",
	}
}

func (s *OAuth2) Attr(name string) (starlark.Value, error) {
	switch name {
	case "id":
		return starlark.String(s.ID), nil

	case "name":
		return starlark.String(s.Name), nil

	case "desc":
		return starlark.String(s.Description), nil

	case "icon":
		return starlark.String(s.Icon), nil

	case "handler":
		if s.StarlarkHandler == nil {
			return starlark.None, nil
		}
		return s.StarlarkHandler, nil

	case "client_id":
		return starlark.String(s.ClientID), nil

	case "authorization_endpoint":
		return starlark.String(s.AuthorizationEndpoint), nil

	case "scopes":
		return s.starlarkScopes, nil

	case "token_endpoint":
		return starlark.String(s.TokenEndpoint), nil

	case "authorization_params":
		if s.starlarkAuthorizationParams == nil {
			return starlark.None, nil
		}
		return s.starlarkAuthorizationParams, nil

	default:
		return nil, nil
	}
}

func (s *OAuth2) String() string       { return "OAuth2(...)" }
func (s *OAuth2) Type() string         { return "OAuth2" }
func (s *OAuth2) Freeze()              {}
func (s *OAuth2) Truth() starlark.Bool { return true }

func (s *OAuth2) Hash() (uint32, error) {
	sum, err := hashstructure.Hash(s, nil)
	return uint32(sum), err
}
