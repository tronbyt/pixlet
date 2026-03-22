package animation

import (
	"github.com/tronbyt/gg"
)

// Translate transforms by translating by a given offset.
type Translate struct {
	// Horizontal offset
	X float64 `starlark:"x,required"`
	// Vertical offset
	Y float64 `starlark:"y,required"`
}

func (t Translate) Apply(ctx *gg.Context, origin Vec2f, rounding Rounding) {
	ctx.Translate(rounding.Apply(t.X), rounding.Apply(t.Y))
}

func (t Translate) Interpolate(other Transform, progress float64) (result Transform, ok bool) {
	if other, ok := other.(Translate); ok {
		return Translate{
			X: Lerp(t.X, other.X, progress),
			Y: Lerp(t.Y, other.Y, progress),
		}, true
	}

	return TranslateDefault, false
}

var TranslateDefault = Translate{X: 0.0, Y: 0.0}
