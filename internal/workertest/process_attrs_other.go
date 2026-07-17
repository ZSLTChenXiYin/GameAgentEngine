//go:build !windows

package workertest

import "os/exec"

func applyProcessAttrs(cmd *exec.Cmd) {}
