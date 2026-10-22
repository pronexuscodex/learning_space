//go:build !(linux || darwin || freebsd || netbsd || openbsd)

package main

import "os/exec"

const runnerSupported = false

func prepareProcess(*exec.Cmd) {}

func cleanupProcess(*exec.Cmd) {}
