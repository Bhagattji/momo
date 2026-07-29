package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyPatchRollbackOnFailure(t *testing.T) {
	d := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(d); err != nil { t.Fatalf("chdir: %v", err) }

	// Create a patch that will succeed adding a file then fail updating a missing file.
	patch := "ACTION: ADD\nPATH: keep.txt\n---\nkeep\n---\n\nACTION: UPDATE\nPATH: missing.txt\n---\nnope\n---\n"
	res, err := ApplyPatch(patch)
	if err == nil {
		t.Fatalf("expected error but got success: %v", res)
	}
	// ensure keep.txt was rolled back (should not exist)
	if _, err := os.Stat("keep.txt"); !os.IsNotExist(err) {
		t.Fatalf("expected keep.txt removed after rollback, exists: %v", err)
	}
}

func TestApplyPatchSymlinkResolution(t *testing.T) {
	d := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(d); err != nil { t.Fatalf("chdir: %v", err) }

	// create a dir outside and symlink into cwd (if symlinks supported)
	outer := t.TempDir()
	if err := os.WriteFile(filepath.Join(outer, "outside.txt"), []byte("x"), 0644); err != nil { t.Fatalf("outer write: %v", err) }
	linkName := "linkdir"
	// try to create symlink; if unsupported, skip
	if err := os.Symlink(outer, linkName); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	// attempt to add file through symlink path which resolves outside workspace
	patch := "ACTION: ADD\nPATH: " + filepath.Join(linkName, "evil.txt") + "\n---\nx\n---\n"
	if _, err := ApplyPatch(patch); err == nil {
		t.Fatalf("expected error for path resolving outside workspace via symlink")
	}
}
