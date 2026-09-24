//go:build !(linux || darwin || freebsd || netbsd || openbsd)

package main

import "errors"

// On other systems (such as Windows) input stays line-based; Ctrl+L
// followed by Enter, or typing "clear", clears the screen instead.

func enableRawInput(fd int) (func(), error) {
	return nil, errors.New("raw input not supported on this platform")
}

func rawInputSupported(fd int) bool { return false }
