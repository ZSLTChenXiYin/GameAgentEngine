package workertest

import (
	"fmt"
	"path/filepath"
	"strings"
)

type EngineFiles struct {
	TempRoot      string
	DBPath        string
	ConfigPath    string
	EngineStdout  string
	EngineStderr  string
	WorkerStdout  string
	WorkerStderr  string
}

func PrepareEngineFiles(tempRoot string) EngineFiles {
	engineStdout, engineStderr := ResolveLogPaths(tempRoot, "engine")
	workerStdout, workerStderr := ResolveLogPaths(tempRoot, "worker")
	return EngineFiles{
		TempRoot:     tempRoot,
		DBPath:       filepath.Join(tempRoot, "gameagentengine.db"),
		ConfigPath:   filepath.Join(tempRoot, "gameagentengine.conf.yaml"),
		EngineStdout: engineStdout,
		EngineStderr: engineStderr,
		WorkerStdout: workerStdout,
		WorkerStderr: workerStderr,
	}
}

func EscapeYAMLPath(path string) string {
	return strings.ReplaceAll(path, "\\", "\\\\")
}

func WriteEngineConfig(path string, content string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("config path is required")
	}
	return WriteFile(path, []byte(content))
}
