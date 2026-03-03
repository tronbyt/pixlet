package random_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/runtime"
)

const randomSrc = `
load("random.star", "random")

min = 100
max = 120

def test_number():
	for x in range(0, 300):
		num = random.number(min, max)
		if num < min:
			fail("random number less than min")
		if num > max:
			fail("random number greater than max")

	# Test overflow protection
	max_int = 9223372036854775807 # math.MaxInt64
	num_overflow = random.number(0, max_int)
	if num_overflow < 0:
		fail("random number overflowed to negative")

	num_max = random.number(max_int, max_int)
	if num_max != max_int:
		fail("random number min/max edge case failed")

def test_seed():
    random.seed(4711)
    sequence = [random.number(0, 1 << 20) for _ in range(500)]

    random.seed(4711) # same seed
    for i in range(len(sequence)):
        if sequence[i] != random.number(0, 1 << 20):
            fail("sequence mismatch despite identical seed")

    random.seed(4712) # different seed
    different = 0
    for i in range(len(sequence)):
        if sequence[i] != random.number(0, 1 << 20):
            different += 1
    if not different:
        fail("sequences identical despite different seeds")

test_number()
test_seed()

def main():
	return []
`

func TestRandom(t *testing.T) {
	app, err := runtime.NewApplet(t.Context(), "random_test.star", []byte(randomSrc), runtime.WithTests(t))
	require.NoError(t, err)

	screens, err := app.Run(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, screens)
}
