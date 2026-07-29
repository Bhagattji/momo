package tools

import (
	"encoding/json"
	"path/filepath"
	"os"
	"strings"
	"testing"
)

func TestRegistryReadFile(t *testing.T) {
	r := NewRegistry()
	tmp := t.TempDir()
	p := filepath.Join(tmp, "a.txt")
	if err := os.WriteFile(p, []byte("abc123"), 0o644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	args := map[string]interface{}{"path": p}
	b, _ := json.Marshal(args)
	res, err := r.Execute("read_file", string(b))
	if err != nil {
		t.Fatalf("registry Execute error: %v", err)
	}
	if !strings.Contains(res.Output, "abc123") {
		t.Fatalf("unexpected output: %q", res.Output)
	}
}
