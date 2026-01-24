package animation

import (
	"github.com/tronbyt/gg"
)

// Translate transforms by translating by a given offset.
//
// DOC(X): Horizontal offset
// DOC(Y): Vertical offset.
type Translate struct {
	Vec2f
}

func (t Translate) Apply(ctx *gg.Context, origin Vec2f, rounding Rounding) {
	ctx.Translate(rounding.Apply(t.X), rounding.Apply(t.Y))
}

func (t Translate) Interpolate(other Transform, progress float64) (result Transform, ok bool) {
	if other, ok := other.(Translate); ok {
		return Translate{t.Lerp(other.Vec2f, progress)}, true
	}

	return TranslateDefault, false
}

var TranslateDefault = Translate{Vec2f{0.0, 0.0}}
