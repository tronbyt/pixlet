package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/url"
	"reflect"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.starlark.net/starlark"
)

const (
	ReturnSchema HandlerReturnType = iota
	ReturnOptions
	ReturnString
	ReturnField
)

const (
	// SchemaFunctionName is the name of the function in Starlark that we expect
	// to be able to call to get the schema for an applet.
	SchemaFunctionName = "get_schema"
)

// Schema holds a configuration object for an applet. It holds a list of fields
// that are exported from an applet.
type Schema struct {
	Version       string         `json:"version"                 validate:"required"`
	Fields        []SchemaField  `json:"schema"                  validate:"dive"`
	Notifications []Notification `json:"notifications,omitempty" validate:"dive"`

	Handlers map[string]SchemaHandler `json:"-"`
}

// SchemaField represents an item in the config used to confgure an applet.
type SchemaField struct {
	Type        string            `json:"type"                  validate:"required,oneof=color datetime dropdown generated location locationbased onoff radio text typeahead oauth2 oauth1 png notification"`
	ID          string            `json:"id"                    validate:"required,excludesall=$"`
	Name        string            `json:"name,omitempty"        validate:"required_for=datetime dropdown location locationbased onoff radio text typeahead png"`
	Description string            `json:"description,omitempty"`
	Icon        string            `json:"icon,omitempty"        validate:"forbidden_for=generated"`
	Visibility  *SchemaVisibility `json:"visibility,omitempty"  validate:"omitempty"`

	Default string         `json:"default,omitempty" validate:"required_for=dropdown onoff radio"`
	Options []SchemaOption `json:"options,omitempty" validate:"required_for=dropdown radio,dive"`
	Palette []string       `json:"palette,omitempty"`
	Sounds  []SchemaSound  `json:"sounds,omitempty"  validate:"required_for=notification,dive"`

	Source          string             `json:"source,omitempty"  validate:"required_for=generated"`
	Handler         string             `json:"handler,omitempty" validate:"required_for=generated locationbased typeahead"`
	StarlarkHandler *starlark.Function `json:"-"`

	// OAuth2 fields come in two modes, checked by validateOAuth2Field. In the
	// original Tidbyt mode the app holds a client ID and a handler that swaps
	// the authorization code for a token. In server-managed mode the app sets
	// a token endpoint instead: the server keeps the OAuth client, signs the
	// user in and passes the app a fresh access token as the field's value.
	ClientID              string            `json:"client_id,omitempty"`
	AuthorizationEndpoint string            `json:"authorization_endpoint,omitempty" validate:"required_for=oauth2"`
	TokenEndpoint         string            `json:"token_endpoint,omitempty"         validate:"allowed_for=oauth2"`
	AuthorizationParams   map[string]string `json:"authorization_params,omitempty"   validate:"allowed_for=oauth2"`
	Scopes                []string          `json:"scopes,omitempty"                 validate:"required_for=oauth2"`

	Secret bool `json:"secret,omitempty" validate:"allowed_for=text"`
}

// SchemaOption represents an option in a field. For example, an item in a drop
// down menu.
type SchemaOption struct {
	Display string `json:"display"`
	Text    string `json:"text"    validate:"required"` // The same as display, for legacy reasons.
	Value   string `json:"value"   validate:"required"`
}

// SchemaSound represents a sound that can be played by the applet.
type SchemaSound struct {
	ID    string `json:"id"    validate:"required"`
	Title string `json:"title" validate:"required"`
	Path  string `json:"path"  validate:"required"`
}

// SchemaVisibility enables conditional fields inside of the mobile app. For
// example, if a field should be invisible until a login is provided.
type SchemaVisibility struct {
	Type      string `json:"type"      validate:"required,oneof=invisible disabled"`
	Condition string `json:"condition" validate:"required,oneof=equal not_equal"`
	Variable  string `json:"variable"  validate:"required"`
	Value     string `json:"value"`
}

// HandlerReturnType defines an enum for the type of information we expect to
// get back from the schema function.
type HandlerReturnType int8

// SchemaHandler defines a function and return type for getting the schema
// for an applet. This can both be the predefined schema function we expect all
// applets to have for config, but can also be used as a callback for.
type SchemaHandler struct {
	Function   *starlark.Function
	ReturnType HandlerReturnType
}

func (s Schema) MarshalJSON() ([]byte, error) {
	type OriginalSchema Schema

	a := struct {
		OriginalSchema
	}{
		OriginalSchema: (OriginalSchema)(s),
	}

	// ensure that fields is serialized as "[]" and not "null",
	// even if there are no fields. otherwise the Tidbyt mobile app breaks
	if a.Fields == nil {
		a.Fields = make([]SchemaField, 0)
	}
	if a.Notifications == nil {
		a.Notifications = make([]Notification, 0)
	}

	js, err := json.Marshal(a)

	return js, err
}

// FromStarlark creates a new Schema from a Starlark schema object.
func FromStarlark(
	val starlark.Value,
	globals starlark.StringDict,
) (*Schema, error) {
	var schema *Schema

	starlarkSchema, ok := val.(*StarlarkSchema)
	if ok {
		schema = &starlarkSchema.Schema
		if schema.Handlers == nil {
			schema.Handlers = make(map[string]SchemaHandler)
			maps.Copy(schema.Handlers, starlarkSchema.Handlers)
		}
	} else {
		// this is a legacy path, where the schema was just a dict
		// instead of a StarlarkSchema object
		schemaTree, err := unmarshalStarlark(val)
		if err != nil {
			return nil, err
		}

		treeJSON, err := json.Marshal(schemaTree)
		if err != nil {
			return nil, err
		}

		schema = &Schema{
			Version:  "1",
			Handlers: make(map[string]SchemaHandler),
		}
		if err := json.Unmarshal(treeJSON, &schema.Fields); err != nil {
			return nil, err
		}
	}

	err := validateSchema(schema)
	if err != nil {
		return nil, err
	}

	for i := range schema.Fields {
		schemaField := &schema.Fields[i]

		var handlerFun *starlark.Function
		if schemaField.StarlarkHandler != nil {
			handlerFun = schemaField.StarlarkHandler
		} else if schemaField.Handler != "" {
			// legacy schema, where the handler was a string instead of
			// a function reference
			handlerValue, ok := globals[schemaField.Handler]
			if !ok {
				return nil, fmt.Errorf(
					"field %d references non-existent handler \"%s\"",
					i,
					schemaField.Handler)
			}

			handlerFun, ok = handlerValue.(*starlark.Function)
			if !ok {
				return nil, fmt.Errorf(
					"field %d references \"%s\" which is not a function",
					i, schemaField.Handler)
			}
		}

		if handlerFun != nil {
			var handlerType HandlerReturnType
			switch schemaField.Type {
			case "locationbased":
				handlerType = ReturnOptions
			case "generated":
				handlerType = ReturnSchema
			case "typeahead":
				handlerType = ReturnOptions
			case "oauth2":
				handlerType = ReturnString
			case "oauth1":
				handlerType = ReturnString
			default:
				return nil, fmt.Errorf(
					"field %d of type \"%s\" can't have a handler function",
					i, schemaField.Type)
			}

			// prepend the field ID to the handler name to avoid conflicts
			schemaField.Handler = fmt.Sprintf("%s$%s", schemaField.ID, schemaField.Handler)
			schema.Handlers[schemaField.Handler] = SchemaHandler{Function: handlerFun, ReturnType: handlerType}
		}
	}

	return schema, nil
}

// EncodeOptions encodes a list of schema options into validated json.
func EncodeOptions(
	starlarkOptions starlark.Value,
) (string, error) {
	optionsTree, err := unmarshalStarlark(starlarkOptions)
	if err != nil {
		return "", err
	}

	options, err := buildOptions(optionsTree)
	if err != nil {
		return "", err
	}

	validate := validator.New()
	for _, o := range options {
		err = validate.Struct(o)
		if err != nil {
			return "", err
		}

		if o.Display == "" {
			o.Display = o.Text
		}
	}

	optionsJson, err := json.Marshal(options)
	if err != nil {
		return "", err
	}

	return string(optionsJson), nil
}

// Transforms a starlark value into Go objects. The value must be a
// list, dict or string. Or a tree of these.
func unmarshalStarlark(object starlark.Value) (any, error) {
	switch v := object.(type) {
	case starlark.String:
		return v.GoString(), nil

	case *starlark.List:
		goList := make([]any, 0, v.Len())

		for listVal := range v.Elements() {
			goVal, err := unmarshalStarlark(listVal)
			if err != nil {
				return nil, err
			}
			goList = append(goList, goVal)
		}

		return goList, nil

	case *starlark.Dict:
		goMap := make(map[string]any)
		for key, value := range v.Entries() {
			strKey, ok := key.(starlark.String)
			if !ok {
				return nil, fmt.Errorf("dict keys must be string")
			}
			goKey := strKey.GoString()

			goVal, err := unmarshalStarlark(value)
			if err != nil {
				return nil, err
			}

			goMap[goKey] = goVal
		}
		return goMap, nil
	case *Option:
		return v.AsSchemaOption(), nil
	case Field:
		return v.AsSchemaField(), nil
	}

	return nil, fmt.Errorf("type %s not allowed in schema", object.Type())
}

// Helper. Verifies that object is string or nil and writes it to *p
// if string.
func setString(p *string, object any) string {
	if object == nil {
		return ""
	}
	objectStr, ok := object.(string)
	if !ok {
		return fmt.Sprintf("expected string, found %T", object)
	}
	*p = objectStr
	return ""
}

func buildOptions(options any) ([]SchemaOption, error) {
	optionsList, ok := options.([]any)
	if !ok {
		return nil, fmt.Errorf("options must be a list")
	}
	schemaOptions := make([]SchemaOption, 0, len(optionsList))
	for j, o := range optionsList {
		op, ok := o.(SchemaOption)
		if ok {
			schemaOptions = append(schemaOptions, op)
			continue
		}

		oMap, ok := o.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("entry %d is not dict or Option", j)
		}

		sop := SchemaOption{}
		if err := setString(&sop.Text, oMap["text"]); err != "" {
			return nil, fmt.Errorf("option %d has bad text: %s", j, err)
		}
		if err := setString(&sop.Value, oMap["value"]); err != "" {
			return nil, fmt.Errorf("option %d has bad value: %s", j, err)
		}

		schemaOptions = append(schemaOptions, sop)
	}
	return schemaOptions, nil
}

func isFieldSet(fl validator.FieldLevel) (bool, bool) {
	switch fl.Field().Kind() {
	case reflect.Map, reflect.Pointer, reflect.Interface:
		return !fl.Field().IsNil(), true
	case reflect.String, reflect.Int, reflect.Slice, reflect.Bool:
		return fl.Field().IsValid() && !fl.Field().IsZero(), true
	default:
		return false, false
	}
}

// Validates a Schema object.
func validateSchema(schema *Schema) error {
	// This custom validator function implements
	// "required_for", which makes the tagged field required
	// whenever SchemaField.Type matches one of the parameters.
	requiredFor := func(fl validator.FieldLevel) bool {
		isSet, ok := isFieldSet(fl)
		if !ok {
			return false
		}
		schemaField := fl.Parent().Interface().(SchemaField)
		if slices.Contains(strings.Split(fl.Param(), " "), schemaField.Type) {
			return isSet
		}
		return true
	}

	// This implements "forbidden_for". Same idea as required for,
	// but the opposite.
	forbiddenFor := func(fl validator.FieldLevel) bool {
		isSet, ok := isFieldSet(fl)
		if !ok {
			return false
		}
		schemaField := fl.Parent().Interface().(SchemaField)
		if slices.Contains(strings.Split(fl.Param(), " "), schemaField.Type) {
			return !isSet
		}
		return true
	}

	allowedFor := func(fl validator.FieldLevel) bool {
		isSet, ok := isFieldSet(fl)
		if !ok {
			return false
		}
		if !isSet {
			// if the field is NOT set, it's fine regardless of type
			return true
		}
		// field IS set — only valid if Type is in the list
		schemaField := fl.Parent().Interface().(SchemaField)
		return slices.Contains(strings.Split(fl.Param(), " "), schemaField.Type)
	}

	// NOTE: It could be helpful to also provide an "optional_for"
	// function, to make sure we catch superfluous tags.

	validate := validator.New()
	_ = validate.RegisterValidation("required_for", requiredFor)
	_ = validate.RegisterValidation("forbidden_for", forbiddenFor)
	_ = validate.RegisterValidation("allowed_for", allowedFor)

	err := validate.Struct(schema)
	if err != nil {
		return err
	}

	for i := range schema.Fields {
		if schema.Fields[i].Type == "oauth2" {
			if err := validateOAuth2Field(&schema.Fields[i]); err != nil {
				return fmt.Errorf("field %q: %w", schema.Fields[i].ID, err)
			}
		}
	}

	return nil
}

// validateOAuth2Field checks that an OAuth2 field is complete for its mode.
func validateOAuth2Field(f *SchemaField) error {
	hasHandler := f.Handler != "" || f.StarlarkHandler != nil
	if f.TokenEndpoint == "" {
		if f.ClientID == "" || !hasHandler {
			return errors.New("an OAuth2 field needs either token_endpoint, or client_id and handler")
		}
		if len(f.AuthorizationParams) > 0 {
			return errors.New("authorization_params requires token_endpoint")
		}
		return nil
	}

	if f.ClientID != "" || hasHandler {
		return errors.New("an OAuth2 field with token_endpoint must not set client_id or handler: the server holds the client and exchanges the code")
	}
	if err := validateOAuth2Endpoint(f.AuthorizationEndpoint); err != nil {
		return fmt.Errorf("authorization_endpoint: %w", err)
	}
	if err := validateOAuth2Endpoint(f.TokenEndpoint); err != nil {
		return fmt.Errorf("token_endpoint: %w", err)
	}
	for k := range f.AuthorizationParams {
		if oauth2ReservedParams[k] {
			return fmt.Errorf("authorization_params must not set %q: the server sets it", k)
		}
	}
	return nil
}

// oauth2ReservedParams are authorization request parameters that the server
// sets itself in server-managed mode.
var oauth2ReservedParams = map[string]bool{
	"client_id":             true,
	"redirect_uri":          true,
	"response_type":         true,
	"scope":                 true,
	"state":                 true,
	"code_challenge":        true,
	"code_challenge_method": true,
}

// validateOAuth2Endpoint accepts absolute https URLs without query, fragment
// or credentials, and plain http on loopback hosts for local testing. Servers
// match OAuth clients to apps by these URLs, so they must be unambiguous.
func validateOAuth2Endpoint(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	switch {
	case u.Opaque != "" || u.Host == "":
		return errors.New("must be an absolute URL")
	case u.User != nil:
		return errors.New("must not contain credentials")
	case u.RawQuery != "" || u.ForceQuery:
		return errors.New("must not contain a query; use authorization_params for extra parameters")
	case u.Fragment != "":
		return errors.New("must not contain a fragment")
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return nil
	case "http":
		host := u.Hostname()
		if ip := net.ParseIP(host); host == "localhost" || (ip != nil && ip.IsLoopback()) {
			return nil
		}
	}
	return errors.New("must use https")
}
