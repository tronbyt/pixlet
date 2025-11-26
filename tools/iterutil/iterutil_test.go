package iterutil

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnumerate(t *testing.T) {
	fn := func() iter.Seq[int] {
		return func(yield func(int) bool) {
			for i := range 10 {
				if !yield(i) {
					return
				}
			}
		}
	}

	for i, v := range Enumerate(fn()) {
		assert.Equal(t, i, v)
	}
}
