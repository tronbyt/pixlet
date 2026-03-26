package strings

import (
	"fmt"

	"github.com/tronbyt/pixlet/starlarkutil"
	"go.starlark.net/starlark"
)

func truncate(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		textParam     starlark.String
		lenParam      starlark.Int
		ellipsisParam starlark.String
	)

	if err := starlark.UnpackArgs(
		fn.Name(), args, kwargs,
		"text", &textParam,
		"length", &lenParam,
		"ellipsis?", &ellipsisParam,
	); err != nil {
		return nil, err
	}

	desiredLen, err := starlarkutil.AsInt[int64](lenParam)
	if err != nil {
		return nil, fmt.Errorf("parsing length: %w", err)
	}

	text := truncateString(string(textParam), string(ellipsisParam), int(desiredLen))
	return starlark.String(text), nil
}

func truncateString(text, ellipsis string, desired int) string {
	desired = max(desired, 0)
	textRunes := []rune(text)
	if len(textRunes) <= desired {
		return text
	}

	ellipsisRunes := []rune(ellipsis)
	if len(ellipsisRunes) == 0 {
		ellipsisRunes = []rune{'…'}
	}
	if len(ellipsisRunes) >= desired {
		return string(ellipsisRunes[:desired])
	}

	return string(textRunes[:desired-len(ellipsisRunes)]) + string(ellipsisRunes)
}
