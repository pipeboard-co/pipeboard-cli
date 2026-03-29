//go:build !windows

package cmd

import (
	"os/exec"
	"syscall"
)

// setSysProcAttr detaches the subprocess into its own process group
// so it survives after the parent exits.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
