//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	procGetConsoleMode = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleMode")
	procSetConsoleMode = syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode")
)

// enableVirtualTerminal turns on escape-sequence processing in the Windows
// console, so colours, clearing the screen and erasing the pager's prompt
// work in the classic console too, not only in Windows Terminal (which
// has it on already). Failure is harmless: the console then shows the
// text as before.
func enableVirtualTerminal() {
	const enableVirtualTerminalProcessing = 0x0004
	h := os.Stdout.Fd()
	var mode uint32
	if r, _, _ := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode))); r == 0 {
		return
	}
	if mode&enableVirtualTerminalProcessing == 0 {
		procSetConsoleMode.Call(h, uintptr(mode|enableVirtualTerminalProcessing))
	}
}
