package workertest

import "testing"

func TestPrepareEngineFiles(t *testing.T) {
	files := PrepareEngineFiles("C:/tmp/demo")
	if files.DBPath == "" || files.ConfigPath == "" || files.EngineStdout == "" || files.WorkerStderr == "" {
		t.Fatalf("unexpected engine files: %#v", files)
	}
}

func TestEscapeYAMLPath(t *testing.T) {
	got := EscapeYAMLPath(`C:\tmp\demo`)
	if got != `C:\\tmp\\demo` {
		t.Fatalf("unexpected escaped path: %s", got)
	}
}
