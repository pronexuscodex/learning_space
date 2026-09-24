//go:build darwin || freebsd || netbsd || openbsd

package main

import "syscall"

// termios ioctl requests on macOS and the BSDs.
const (
	ioctlGetTermios = syscall.TIOCGETA
	ioctlSetTermios = syscall.TIOCSETA
)
