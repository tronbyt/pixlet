package main

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func must2[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
