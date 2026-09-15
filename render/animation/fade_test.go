package animation

import (
	"testing"
)

func assertInterpolateFade(
	t *testing.T,
	expected float64,
	from float64,
	to float64,
	progress float64,
) {
	AssertInterpolate(t, Fade{Value: expected}, Fade{Value: from}, Fade{Value: to}, progress)
}

func TestInterpolateFade(t *testing.T) {
	from := 0.0
	to := 1.0

	assertInterpolateFade(t, 0.0, from, to, 0.0)
	assertInterpolateFade(t, 0.1, from, to, 0.1)
	assertInterpolateFade(t, 0.5, from, to, 0.5)
	assertInterpolateFade(t, 1.0, from, to, 1.0)
}
