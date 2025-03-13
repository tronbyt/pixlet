//go:build windows

package encode

import "syscall"

func loadLibrary(library string) (uintptr, error) {
	lib, err := syscall.LoadLibrary(library)
	if err != nil {
		return 0, err
	}

	return uintptr(lib), nil
}
