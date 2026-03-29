//go:build windows

package cmd

import "os/exec"

// setSysProcAttr is a no-op on Windows; the subprocess runs independently
// by default when stdout/stderr are nil.
func setSysProcAttr(cmd *exec.Cmd) {}
