package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveProjectAndLoadMerge(t *testing.T) {
	wd := t.TempDir()
	// create project config
	cfg := &Config{DefaultProvider: "groq", DefaultModel: "m", Approvals: map[string]bool{"run_cmd": true}}
	if err := SaveProject(cfg, wd); err != nil {
		t.Fatalf("SaveProject failed: %v", err)
	}
	// Change working dir to a subdir to ensure Load walks up
	sub := filepath.Join(wd, "subdir")
	if err := os.MkdirAll(sub, 0o755); err != nil { t.Fatalf("mkdir: %v", err) }
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(sub); err != nil { t.Fatalf("chdir: %v", err) }
	lcfg, err := Load()
	if err != nil { t.Fatalf("Load failed: %v", err) }
	if !lcfg.Approvals["run_cmd"] {
		t.Fatalf("project approval not merged into Load config")
	}
}
