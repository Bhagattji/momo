package tools

import (
	"path/filepath"
	"os"
	"testing"
)

func TestReadFile(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "test.txt")
	content := "hello world"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	res, err := ReadFile(p, 0, 0)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if res.IsError {
		t.Fatalf("ReadFile result is error: %s", res.Output)
	}
	if res.Output != content {
		t.Fatalf("unexpected content: %q", res.Output)
	}
}
