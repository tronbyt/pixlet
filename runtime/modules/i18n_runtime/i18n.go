package i18n_runtime

import (
	"fmt"
	"sync"

	"github.com/qri-io/starlib/util"
	"go.starlark.net/starlark"
	"golang.org/x/text/message"
)

const ModuleName = "i18n.star"

var (
	once   sync.Once
	module starlark.StringDict
)

func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		module = starlark.StringDict{
			"tr": starlark.NewBuiltin("tr", Translate),
		}
	})
	return module, nil
}

var (
	ErrUnexpectedKwargs = fmt.Errorf("unexpected keyword arguments")
	ErrMissingFormatArg = fmt.Errorf("missing format argument")
	ErrInvalidFormatArg = fmt.Errorf("format argument must be a string")
	ErrInvalidParam     = fmt.Errorf("invalid param")
)

func Translate(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(kwargs) > 0 {
		return nil, ErrUnexpectedKwargs
	}
	if len(args) == 0 {
		return nil, ErrMissingFormatArg
	}

	format, ok := args[0].(starlark.String)
	if !ok {
		return nil, fmt.Errorf("%w, got %s", ErrInvalidFormatArg, args[0].Type())
	}

	params := args[1:]

	goArgs := make([]any, len(params))
	for i, v := range params {
		var err error
		goArgs[i], err = util.Unmarshal(v)
		if err != nil {
			return nil, fmt.Errorf("%w %d: %w", ErrInvalidParam, i, err)
		}
	}

	lang := LanguageFromThread(thread)
	catalog := CatalogFromThread(thread)
	matched, _, _ := catalog.Matcher().Match(lang)
	printer := message.NewPrinter(matched, message.Catalog(catalog))
	result := printer.Sprintf(string(format), goArgs...)
	return starlark.String(result), nil
}
