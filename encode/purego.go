//go:build !windows

package encode

import "github.com/ebitengine/purego"

func loadLibrary(library string) (uintptr, error) {
	return purego.Dlopen(library, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}
