//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var procGetConsoleScreenBufferInfo = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleScreenBufferInfo")

// terminalSize asks the Windows console for its visible window size.
func terminalSize(fd int) (cols, rows int, ok bool) {
	type coord struct{ X, Y int16 }
	var info struct {
		Size, Cursor             coord
		Attributes               uint16
		Left, Top, Right, Bottom int16
		MaxSize                  coord
	}
	r, _, _ := procGetConsoleScreenBufferInfo.Call(uintptr(fd), uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return 0, 0, false
	}
	return int(info.Right-info.Left) + 1, int(info.Bottom-info.Top) + 1, true
}
