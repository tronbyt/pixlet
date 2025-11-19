package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/unicode/bidi"
)

func TestBaseDirection(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bidi.Direction
	}{
		{name: "ltr text", args: args{s: "Hello, world"}, want: bidi.LeftToRight},
		{name: "rtl text", args: args{s: "שלום"}, want: bidi.RightToLeft},
		{name: "leading neutrals before rtl", args: args{s: "123 שלום"}, want: bidi.RightToLeft},
		{name: "no strong directionals", args: args{s: "12345!?."}, want: bidi.LeftToRight},
		{name: "invalid utf8 falls back to ltr", args: args{s: string([]byte{0xff, 0xfe})}, want: bidi.LeftToRight},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BaseDirection(tt.args.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVisualBidiString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty string", args: args{s: ""}, want: ""},
		{name: "pure ltr", args: args{s: "Pixlet"}, want: "Pixlet"},
		{name: "pure rtl reverses", args: args{s: "שלום"}, want: "םולש"},
		{name: "ltr paragraph with rtl run", args: args{s: "abc שלום def"}, want: "abc םולש def"},
		{name: "rtl paragraph with trailing ltr", args: args{s: "שלום abc"}, want: "abc םולש"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VisualBidiString(tt.args.s)
			assert.Equal(t, tt.want, got)
		})
	}
}
