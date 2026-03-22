package animation

import (
	"testing"
)

func assertInterpolateTranslate(
	t *testing.T,
	expected Translate,
	from Translate,
	to Translate,
	progress float64,
) {
	AssertInterpolate(t, expected, from, to, progress)
}

func TestInterpolateTranslate(t *testing.T) {
	from := Translate{X: 0.0, Y: 0.0}
	to := Translate{X: 100.0, Y: 200.0}

	assertInterpolateTranslate(t, Translate{X: 0.0, Y: 0.0}, from, to, 0.0)
	assertInterpolateTranslate(t, Translate{X: 10.0, Y: 20.0}, from, to, 0.1)
	assertInterpolateTranslate(t, Translate{X: 33.0, Y: 66.0}, from, to, 0.33)
	assertInterpolateTranslate(t, Translate{X: 100.0, Y: 200.0}, from, to, 1.0)
	assertInterpolateTranslate(t, Translate{X: 1337.0, Y: 2674.0}, from, to, 13.37)
}
