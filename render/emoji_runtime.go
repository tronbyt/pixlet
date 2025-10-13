package render

//go:generate go run ./gen/emoji_pack.go

import (
	"strings"

	"github.com/rivo/uniseg"
	"tidbyt.dev/pixlet/fonts/emoji"
)

// segmentEmoji breaks a string into a sequence of tokens, where each token is either
// an emoji sequence key present in Index, or a plain text segment (no emoji inside).
// Segments are identified using Unicode grapheme clusters so complex emoji remain intact.
func segmentEmoji(s string) ([]segment, bool) {
	var hasEmoji bool
	segments := make([]segment, 0, 1)
	var buf strings.Builder
	buf.Grow(len(s))

	state := -1
	for len(s) != 0 {
		var cluster string
		cluster, s, _, state = uniseg.FirstGraphemeClusterInString(s, state)
		if _, ok := emoji.Index[cluster]; ok {
			if buf.Len() != 0 {
				segments = append(segments, segment{text: buf.String()})
				buf.Reset()
			}
			hasEmoji = true
			segments = append(segments, segment{
				emoji: true,
				text:  cluster,
			})
		} else {
			buf.WriteString(cluster)
		}
	}

	if buf.Len() != 0 {
		segments = append(segments, segment{text: buf.String()})
	}

	return segments, hasEmoji
}

type segment struct {
	emoji bool
	text  string // either a plain text run or the exact emoji sequence string
}
