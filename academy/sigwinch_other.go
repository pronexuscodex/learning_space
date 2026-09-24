//go:build !(linux || darwin || freebsd || netbsd || openbsd)

package main

import "os"

// watchResize: this platform does not signal resizes, so layouts adapt on
// the next screen (or on Ctrl+L) instead.
func watchResize() <-chan os.Signal { return nil }
