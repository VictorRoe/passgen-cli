//go:build windows

package session

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// detach starts the agent without a console and outside the terminal's
// process group, so closing the terminal or Ctrl+C does not kill it.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
		HideWindow:    true,
	}
}
