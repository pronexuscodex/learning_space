//go:build linux || darwin || freebsd || netbsd || openbsd

package main

import (
	"os"
	"os/signal"
	"syscall"
)

// watchResize delivers a notification each time the terminal is resized.
func watchResize() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	return ch
}
