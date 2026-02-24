package colorutil

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromHex(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name    string
		args    args
		want    color.Color
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "#RGB",
			args:    args{text: "#5ad"},
			want:    color.NRGBA{R: 0x55, G: 0xAA, B: 0xDD, A: 0xFF},
			wantErr: require.NoError,
		},
		{
			name:    "RGB",
			args:    args{text: "5ad"},
			want:    color.NRGBA{R: 0x55, G: 0xAA, B: 0xDD, A: 0xFF},
			wantErr: require.NoError,
		},
		{
			name:    "#RGBA",
			args:    args{text: "#5ad8"},
			want:    color.NRGBA{R: 0x55, G: 0xAA, B: 0xDD, A: 0x88},
			wantErr: require.NoError,
		},
		{
			name:    "RGBA",
			args:    args{text: "5ad8"},
			want:    color.NRGBA{R: 0x55, G: 0xAA, B: 0xDD, A: 0x88},
			wantErr: require.NoError,
		},
		{
			name:    "#RRGGBB",
			args:    args{text: "#257adb"},
			want:    color.NRGBA{R: 0x25, G: 0x7A, B: 0xDB, A: 0xFF},
			wantErr: require.NoError,
		},
		{
			name:    "RRGGBB",
			args:    args{text: "257adb"},
			want:    color.NRGBA{R: 0x25, G: 0x7A, B: 0xDB, A: 0xFF},
			wantErr: require.NoError,
		},
		{
			name:    "#RRGGBBAA",
			args:    args{text: "#257adb75"},
			want:    color.NRGBA{R: 0x25, G: 0x7A, B: 0xDB, A: 0x75},
			wantErr: require.NoError,
		},
		{
			name:    "RRGGBBAA",
			args:    args{text: "257adb75"},
			want:    color.NRGBA{R: 0x25, G: 0x7A, B: 0xDB, A: 0x75},
			wantErr: require.NoError,
		},
		{
			name:    "invalid length",
			args:    args{text: "#f0"},
			want:    color.NRGBA{},
			wantErr: require.Error,
		},
		{
			name:    "invalid char",
			args:    args{text: "#zzz"},
			want:    color.NRGBA{},
			wantErr: require.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseHex(tt.args.text)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToHex(t *testing.T) {
	type args struct {
		c color.NRGBA
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "RGB",
			args: args{color.NRGBA{R: 0x55, G: 0xAA, B: 0xDD, A: 0xFF}},
			want: "#5ad",
		},
		{
			name: "RGBA",
			args: args{color.NRGBA{R: 0x55, G: 0xAA, B: 0xDD, A: 0x88}},
			want: "#5ad8",
		},
		{
			name: "RRGGBB",
			args: args{color.NRGBA{R: 0x25, G: 0x7A, B: 0xDB, A: 0xFF}},
			want: "#257adb",
		},
		{
			name: "RRGGBBAA",
			args: args{color.NRGBA{R: 0x25, G: 0x7A, B: 0xDB, A: 0x75}},
			want: "#257adb75",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatHex(tt.args.c))
		})
	}
}
