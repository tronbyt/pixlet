package canvas

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMaxDurationMillis(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   time.Duration
		want int
	}{
		{"unset means unbounded", 0, 0},
		{"whole seconds", 15 * time.Second, 15000},
		{"sub-second", 1500 * time.Millisecond, 1500},
		{"truncates toward zero", 1999 * time.Microsecond, 1},
		{"below a millisecond", 999 * time.Microsecond, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, Metadata{MaxDuration: tc.in}.MaxDurationMillis())
		})
	}
}
