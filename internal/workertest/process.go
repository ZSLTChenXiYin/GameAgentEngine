package workertest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type ManagedProcess struct {
	Cmd       *exec.Cmd
	StdoutLog string
	StderrLog string
}

func StartProcess(filePath string, args []string, workingDir string, stdoutPath string, stderrPath string) (*ManagedProcess, error) {
	if filePath == "" {
		return nil, fmt.Errorf("process file path is required")
	}
	if stdoutPath == "" || stderrPath == "" {
		return nil, fmt.Errorf("stdout/stderr log paths are required")
	}
	stdoutFile, err := os.Create(stdoutPath)
	if err != nil {
		return nil, err
	}
	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		_ = stdoutFile.Close()
		return nil, err
	}
	cmd := exec.Command(filePath, args...)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	applyProcessAttrs(cmd)
	if err := cmd.Start(); err != nil {
		_ = stdoutFile.Close()
		_ = stderrFile.Close()
		return nil, err
	}
	_ = stdoutFile.Close()
	_ = stderrFile.Close()
	return &ManagedProcess{Cmd: cmd, StdoutLog: stdoutPath, StderrLog: stderrPath}, nil
}

func StopProcess(proc *ManagedProcess) error {
	if proc == nil || proc.Cmd == nil || proc.Cmd.Process == nil {
		return nil
	}
	if proc.Cmd.ProcessState != nil && proc.Cmd.ProcessState.Exited() {
		return nil
	}
	return proc.Cmd.Process.Kill()
}

func ResolveLogPaths(tempRoot string, prefix string) (string, string) {
	return filepath.Join(tempRoot, prefix+".stdout.log"), filepath.Join(tempRoot, prefix+".stderr.log")
}
