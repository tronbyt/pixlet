package emoji

import (
	"reflect"
	"testing"
)

func TestSegmentString(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		want         []Segment
		wantHasEmoji bool
	}{
		{
			name:  "plain text only",
			input: "Hello world",
			want: []Segment{
				{Text: "Hello world"},
			},
			wantHasEmoji: false,
		},
		{
			name:  "single emoji",
			input: "😀",
			want: []Segment{
				{Text: "😀", IsEmoji: true},
			},
			wantHasEmoji: true,
		},
		{
			name:  "emoji with surrounding text",
			input: "Hello 😀 World",
			want: []Segment{
				{Text: "Hello "},
				{Text: "😀", IsEmoji: true},
				{Text: " World"},
			},
			wantHasEmoji: true,
		},
		{
			name:  "consecutive emoji",
			input: "😀😂😍",
			want: []Segment{
				{Text: "😀", IsEmoji: true},
				{Text: "😂", IsEmoji: true},
				{Text: "😍", IsEmoji: true},
			},
			wantHasEmoji: true,
		},
		{
			name:  "flag emoji sequence",
			input: "🇺🇸",
			want: []Segment{
				{Text: "🇺🇸", IsEmoji: true},
			},
			wantHasEmoji: true,
		},
		{
			name:  "unknown emoji stays in text",
			input: "Hi 🤷‍♀️ there",
			want: []Segment{
				{Text: "Hi 🤷‍♀️ there"},
			},
			wantHasEmoji: false,
		},
		{
			name:  "unknown then known emoji",
			input: "🤷‍♀️😀a",
			want: []Segment{
				{Text: "🤷‍♀️"},
				{Text: "😀", IsEmoji: true},
				{Text: "a"},
			},
			wantHasEmoji: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, hasEmoji := SegmentString(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SegmentString() got = %v, want %v", got, tt.want)
			}
			if hasEmoji != tt.wantHasEmoji {
				t.Fatalf("SegmentString(%q) hasEmoji = %v, want %v", tt.input, hasEmoji, tt.wantHasEmoji)
			}
		})
	}
}
