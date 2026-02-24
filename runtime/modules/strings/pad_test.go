package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPadString(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		char     string
		desired  int
		align    PadAlign
		expected string
	}{
		{
			name:     "pad empty char uses space",
			text:     "foo",
			char:     "",
			desired:  5,
			align:    AlignStart,
			expected: "foo  ",
		},
		{
			name:     "pad with specific char",
			text:     "foo",
			char:     "-",
			desired:  5,
			align:    AlignStart,
			expected: "foo--",
		},
		{
			name:     "no padding needed",
			text:     "foobar",
			desired:  5,
			align:    AlignStart,
			expected: "foobar",
		},
		{
			name:     "pad align end",
			text:     "foo",
			char:     ".",
			desired:  5,
			align:    AlignEnd,
			expected: "..foo",
		},
		{
			name:     "pad with multi-char string",
			text:     "foo",
			char:     "ab",
			desired:  8,
			align:    AlignStart,
			expected: "fooababa",
		},
		{
			name:     "pad with multi-char string truncated",
			text:     "foo",
			char:     "ab",
			desired:  6,
			align:    AlignStart,
			expected: "fooaba",
		},
		{
			name:     "pad align end with multi-char string",
			text:     "foo",
			char:     "ab",
			desired:  8,
			align:    AlignEnd,
			expected: "ababafoo",
		},
		{
			name:     "pad with emoji char",
			text:     "hi",
			char:     "✨",
			desired:  5,
			align:    AlignStart,
			expected: "hi✨✨✨",
		},
		{
			name:     "pad text containing emojis",
			text:     "👋",
			char:     ".",
			desired:  3,
			align:    AlignEnd,
			expected: "..👋",
		},
		{
			name:     "pad with multi-byte characters",
			text:     "世界",
			char:     "！",
			desired:  5,
			align:    AlignStart,
			expected: "世界！！！",
		},
		{
			name:     "pad to zero",
			text:     "foo",
			char:     " ",
			desired:  0,
			align:    AlignStart,
			expected: "foo",
		},
		{
			name:     "negative length",
			text:     "foo",
			char:     " ",
			desired:  -1,
			align:    AlignStart,
			expected: "foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := padString(tt.text, tt.char, tt.desired, tt.align)
			assert.Equal(t, tt.expected, got)
		})
	}
}
