package workertest

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func MakeTempRoot(prefix string) (string, error) {
	stamp := time.Now().Format("20060102150405")
	baseDir := filepath.Join("tmp", "worker-tests")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return "", err
	}
	return os.MkdirTemp(baseDir, fmt.Sprintf("%s-%s-", prefix, stamp))
}

func RemoveTempRoot(path string, keep bool) error {
	if keep || path == "" {
		return nil
	}
	return os.RemoveAll(path)
}

func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
