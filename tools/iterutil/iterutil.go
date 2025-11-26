package iterutil

import "iter"

// Enumerate lifts an iter.Seq to an iter.Seq2 with an index
func Enumerate[T any](s iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		s(func(v T) bool {
			ok := yield(i, v)
			i++
			return ok
		})
	}
}
