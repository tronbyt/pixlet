package emoji

import (
	"reflect"
	"testing"

	font "tidbyt.dev/pixlet/fonts/emoji"
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
			name:  "numbers",
			input: "1234567890",
			want: []Segment{
				{Text: "1234567890"},
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
			name:  "plain arrow",
			input: "↗",
			want: []Segment{
				{Text: "↗"},
			},
			wantHasEmoji: false,
		},
		{
			name:  "emoji arrow",
			input: "↗️",
			want: []Segment{
				{Text: "↗️", IsEmoji: true},
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

func TestGetUsesFallback(t *testing.T) {
	if font.Fallback.Empty() {
		t.Fatal("fallback sprite is empty")
	}
	img, exists, err := Get("not-a-real-emoji", false)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if exists {
		t.Fatalf("Get() returned true for non-existent emoji")
	}
	got := img.Bounds()
	want := font.Fallback
	if got.Dx() != want.Dx() || got.Dy() != want.Dy() {
		t.Fatalf("fallback image has bounds %v, want %v", got, want)
	}
}
