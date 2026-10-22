//go:build linux || darwin || freebsd || netbsd || openbsd

package main

import (
	"os/exec"
	"syscall"
)

const runnerSupported = true

// prepareProcess puts the program in its own process group, so that the
// time limit kills it together with anything it started.
func prepareProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
}

// cleanupProcess kills whatever the program left running in its group.
func cleanupProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
