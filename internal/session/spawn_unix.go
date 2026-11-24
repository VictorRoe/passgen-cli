//go:build !windows

package session

import (
	"os/exec"
	"syscall"
)

// detach starts the agent in its own session so it outlives the terminal.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
