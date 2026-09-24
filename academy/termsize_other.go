//go:build !(linux || darwin || freebsd || netbsd || openbsd || windows)

package main

// terminalSize is unknown on this platform; callers fall back to
// $COLUMNS/$LINES or 80×24.
func terminalSize(fd int) (cols, rows int, ok bool) { return 0, 0, false }
