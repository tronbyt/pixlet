package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		ellipsis string
		desired  int
		expected string
	}{
		{
			name:     "no truncation needed",
			text:     "hello",
			desired:  10,
			expected: "hello",
		},
		{
			name:     "exact length no truncation",
			text:     "hello",
			desired:  5,
			expected: "hello",
		},
		{
			name:     "truncate with default ellipsis",
			text:     "hello world",
			desired:  5,
			expected: "hell…",
		},
		{
			name:     "truncate with custom ellipsis",
			text:     "hello world",
			ellipsis: "...",
			desired:  5,
			expected: "he...",
		},
		{
			name:     "truncate empty string",
			text:     "",
			desired:  5,
			expected: "",
		},
		{
			name:     "truncate unicode",
			text:     "世界こんにちは",
			desired:  3,
			expected: "世界…",
		},
		{
			name:     "truncate with emoji ellipsis",
			text:     "too long",
			ellipsis: "✨",
			desired:  4,
			expected: "too✨",
		},
		{
			name:     "truncate ellipsis longer than desired",
			text:     "hello world",
			ellipsis: "...",
			desired:  2,
			expected: "..",
		},
		{
			name:     "truncate ellipsis equal to desired",
			text:     "hello world",
			ellipsis: "...",
			desired:  3,
			expected: "...",
		},
		{
			name:     "truncate to zero",
			text:     "hello world",
			ellipsis: "...",
			desired:  0,
			expected: "",
		},
		{
			name:     "negative length",
			text:     "hello world",
			ellipsis: "...",
			desired:  -1,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.text, tt.ellipsis, tt.desired)
			assert.Equal(t, tt.expected, got)
		})
	}
}
