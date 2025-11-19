package i18n

import (
	"strings"

	"golang.org/x/text/unicode/bidi"
)

func BaseDirection(s string) bidi.Direction {
	for p := 0; p < len(s); {
		props, sz := bidi.LookupString(s[p:])
		switch props.Class() {
		case bidi.L, bidi.LRE, bidi.LRO:
			return bidi.LeftToRight
		case bidi.R, bidi.AL, bidi.RLE, bidi.RLO:
			return bidi.RightToLeft
		}
		if sz > 0 {
			p += sz
		} else {
			// Advance by one byte to avoid infinite loop on invalid UTF-8.
			p++
		}
	}
	// Fallback if no strong chars found
	return bidi.LeftToRight
}

func VisualBidiString(s string) string {
	if s == "" {
		return s
	}

	var p bidi.Paragraph
	if _, err := p.SetString(s); err != nil {
		return s
	}

	order, err := p.Order()
	if err != nil {
		return s
	}

	dir := order.Direction()
	var out strings.Builder
	out.Grow(len(s))

	for i := range order.NumRuns() {
		if dir == bidi.RightToLeft {
			i = order.NumRuns() - 1 - i
		}
		run := order.Run(i)
		if run.Direction() == bidi.RightToLeft {
			out.WriteString(bidi.ReverseString(run.String()))
		} else {
			out.WriteString(run.String())
		}
	}

	return out.String()
}
