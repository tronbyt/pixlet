package animation

import (
	"testing"
)

func assertInterpolateScale(
	t *testing.T,
	expected Scale,
	from Scale,
	to Scale,
	progress float64,
) {
	AssertInterpolate(t, expected, from, to, progress)
}

func TestInterpolateScale(t *testing.T) {
	from := Scale{X: 0.0, Y: 0.0}
	to := Scale{X: 100.0, Y: 200.0}

	assertInterpolateScale(t, Scale{X: 0.0, Y: 0.0}, from, to, 0.0)
	assertInterpolateScale(t, Scale{X: 10.0, Y: 20.0}, from, to, 0.1)
	assertInterpolateScale(t, Scale{X: 33.0, Y: 66.0}, from, to, 0.33)
	assertInterpolateScale(t, Scale{X: 100.0, Y: 200.0}, from, to, 1.0)
	assertInterpolateScale(t, Scale{X: 1337.0, Y: 2674.0}, from, to, 13.37)
}
