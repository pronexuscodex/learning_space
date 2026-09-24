//go:build linux || darwin || freebsd || netbsd || openbsd

package main

import (
	"syscall"
	"unsafe"
)

// terminalSize asks the kernel for the terminal's current size.
func terminalSize(fd int) (cols, rows int, ok bool) {
	var ws struct{ Row, Col, Xpixel, Ypixel uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Col == 0 {
		return 0, 0, false
	}
	return int(ws.Col), int(ws.Row), true
}
