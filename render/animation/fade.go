package animation

import (
	"image"
	"math"

	"github.com/tronbyt/gg"
)

// Fade transforms by adjusting the opacity of the child.
//
// A value of `0` renders the child fully transparent, `1` renders it fully
// opaque, and values in between blend it proportionally.
//
// Note: nesting a `Fade` inside another `Fade` doesn't combine them the way
// you might expect. Only the innermost value takes effect, rather than the
// two fade amounts multiplying together (e.g. two 50% fades won't add up to
// 25% visible). Keep fades at the same level rather than nesting them.
type Fade struct {
	// Opacity to fade to, from 0.0 (fully transparent) to 1.0 (fully opaque).
	Value float64 `starlark:"value,required"`
}

func (f Fade) Apply(ctx *gg.Context, origin Vec2f, rounding Rounding) {
	alpha := uint8(math.Round(math.Max(0.0, math.Min(1.0, f.Value)) * 255))

	mask := image.NewAlpha(ctx.Image().Bounds())
	for i := range mask.Pix {
		mask.Pix[i] = alpha
	}

	// The mask is built from ctx's own bounds, so this cannot fail.
	_ = ctx.SetMask(mask)
}

func (f Fade) Interpolate(other Transform, progress float64) (result Transform, ok bool) {
	if other, ok := other.(Fade); ok {
		return Fade{
			Value: Lerp(f.Value, other.Value, progress),
		}, true
	}

	return FadeDefault, false
}

var FadeDefault = Fade{Value: 1.0}
