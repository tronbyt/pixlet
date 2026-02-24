package colorutil

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromHSV(t *testing.T) {
	type args struct {
		h float64
		s float64
		v float64
		a uint8
	}
	tests := []struct {
		name string
		args args
		want color.Color
	}{
		{"black", args{0, 0, 0, 255}, color.NRGBA{A: 0xFF}},
		{"white", args{0, 0, 1, 255}, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}},
		{"red", args{0, 1, 1, 255}, color.NRGBA{R: 0xFF, A: 0xFF}},
		{"green", args{120, 1, 1, 255}, color.NRGBA{G: 0xFF, A: 0xFF}},
		{"blue", args{240, 1, 1, 255}, color.NRGBA{B: 0xFF, A: 0xFF}},
		{"yellow", args{60, 1, 1, 255}, color.NRGBA{R: 0xFF, G: 0xFF, A: 0xFF}},
		{"cyan", args{180, 1, 1, 255}, color.NRGBA{G: 0xFF, B: 0xFF, A: 0xFF}},
		{"magenta", args{300, 1, 1, 255}, color.NRGBA{R: 0xFF, B: 0xFF, A: 0xFF}},
		{"gray", args{0, 0, 0.5, 255}, color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xFF}}, // approx
		{"red at 50%", args{0, 1, 0.5, 255}, color.NRGBA{R: 0x80, A: 0xFF}},
		{"red wrapped", args{360, 1, 1, 255}, color.NRGBA{R: 0xFF, A: 0xFF}},
		{"red wrapped 2", args{720, 1, 1, 255}, color.NRGBA{R: 0xFF, A: 0xFF}},
		{"red negative", args{-360, 1, 1, 255}, color.NRGBA{R: 0xFF, A: 0xFF}},
		{"cyan wrapped", args{540, 1, 1, 255}, color.NRGBA{G: 0xFF, B: 0xFF, A: 0xFF}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseHSV(tt.args.h, tt.args.s, tt.args.v, tt.args.a)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToHSV(t *testing.T) {
	type args struct {
		c color.NRGBA
	}
	tests := []struct {
		name  string
		args  args
		wantH float64
		wantS float64
		wantV float64
		wantA uint8
	}{
		{"black", args{color.NRGBA{A: 0xFF}}, 0, 0, 0, 255},
		{"white", args{color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}}, 0, 0, 1, 255},
		{"red", args{color.NRGBA{R: 0xFF, A: 0xFF}}, 0, 1, 1, 255},
		{"green", args{color.NRGBA{G: 0xFF, A: 0xFF}}, 120, 1, 1, 255},
		{"blue", args{color.NRGBA{B: 0xFF, A: 0xFF}}, 240, 1, 1, 255},
		{"yellow", args{color.NRGBA{R: 0xFF, G: 0xFF, A: 0xFF}}, 60, 1, 1, 255},
		{"cyan", args{color.NRGBA{G: 0xFF, B: 0xFF, A: 0xFF}}, 180, 1, 1, 255},
		{"magenta", args{color.NRGBA{R: 0xFF, B: 0xFF, A: 0xFF}}, 300, 1, 1, 255},
		{"gray", args{color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xFF}}, 0, 0, 0.5, 255}, // approx
		{"red at 50%", args{color.NRGBA{R: 0x80, A: 0xFF}}, 0, 1, 0.5, 255},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotH, gotS, gotV, gotA := FormatHSV(tt.args.c)
			assert.InDelta(t, tt.wantH, gotH, 1e-6)
			assert.InDelta(t, tt.wantS, gotS, 0.01)
			assert.InDelta(t, tt.wantV, gotV, 0.01)
			assert.Equal(t, tt.wantA, gotA)
		})
	}
}
