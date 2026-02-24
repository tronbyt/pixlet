package strings

import (
	"fmt"

	"github.com/tronbyt/pixlet/starlarkutil"
	"go.starlark.net/starlark"
)

//go:generate go tool enumer -type PadAlign -trimprefix Align -transform lower

type PadAlign uint8

const (
	AlignStart PadAlign = iota
	AlignEnd
)

func pad(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		textParam  starlark.String
		lenParam   starlark.Int
		alignParam starlark.String
		charParam  starlark.String
	)

	if err := starlark.UnpackArgs(
		fn.Name(), args, kwargs,
		"text", &textParam,
		"length", &lenParam,
		"align?", &alignParam,
		"char?", &charParam,
	); err != nil {
		return nil, err
	}

	desiredLen, err := starlarkutil.AsInt64(lenParam)
	if err != nil {
		return nil, fmt.Errorf("parsing length: %w", err)
	}

	var align PadAlign
	if alignParam != "" {
		if align, err = PadAlignString(string(alignParam)); err != nil {
			return nil, fmt.Errorf("parsing align: %w", err)
		}
	}

	text := padString(string(textParam), string(charParam), int(desiredLen), align)
	return starlark.String(text), nil
}

func padString(text, char string, desired int, align PadAlign) string {
	textRunes := []rune(text)
	if len(textRunes) >= desired {
		return text
	}

	desired = max(0, min(desired, 512))
	paddingNeeded := desired - len(textRunes)
	p := make([]rune, 0, paddingNeeded)
	charRunes := []rune(char)

	if len(charRunes) == 0 {
		charRunes = []rune{' '}
	}

	for len(p) < paddingNeeded {
		p = append(p, charRunes...)
	}
	p = p[:paddingNeeded]

	if align == AlignEnd {
		return string(p) + string(textRunes)
	}
	return string(textRunes) + string(p)
}
