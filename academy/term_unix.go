//go:build linux || darwin || freebsd || netbsd || openbsd

package main

// Raw keyboard mode via termios, using only the standard library. Only
// line buffering (ICANON) and echo are switched off; signal keys such as
// Ctrl-C still work and output processing is untouched, so everything
// else behaves normally.

import (
	"syscall"
	"unsafe"
)

func getTermios(fd int) (syscall.Termios, error) {
	var t syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(ioctlGetTermios), uintptr(unsafe.Pointer(&t)))
	if errno != 0 {
		return t, errno
	}
	return t, nil
}

func setTermios(fd int, t *syscall.Termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(ioctlSetTermios), uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return errno
	}
	return nil
}

// enableRawInput switches the terminal to key-at-a-time input without echo
// and returns a function that restores the previous settings.
func enableRawInput(fd int) (func(), error) {
	old, err := getTermios(fd)
	if err != nil {
		return nil, err
	}
	raw := old
	raw.Lflag &^= syscall.ICANON | syscall.ECHO | syscall.IEXTEN
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := setTermios(fd, &raw); err != nil {
		return nil, err
	}
	return func() { _ = setTermios(fd, &old) }, nil
}

// rawInputSupported reports whether key-at-a-time input can work here.
func rawInputSupported(fd int) bool {
	_, err := getTermios(fd)
	return err == nil
}
