//go:build !windows

package main

// enableVirtualTerminal is needed only on Windows; other terminals handle
// escape sequences already.
func enableVirtualTerminal() {}
