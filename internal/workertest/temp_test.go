package workertest

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMakeTempRootUsesWorkspaceTmp(t *testing.T) {
	path, err := MakeTempRoot("gae-test")
	if err != nil {
		t.Fatalf("MakeTempRoot returned error: %v", err)
	}
	defer RemoveTempRoot(path, false)
	normalized := filepath.ToSlash(path)
	if !strings.Contains(normalized, "tmp/worker-tests/") {
		t.Fatalf("expected temp root under tmp/worker-tests, got %q", path)
	}
}
