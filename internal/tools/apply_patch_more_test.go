package tools

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestApplyPatchMultipleActions(t *testing.T) {
	d := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(d); err != nil { t.Fatalf("chdir: %v", err) }

	patch := "ACTION: ADD\nPATH: a.txt\n---\nA\n---\n\nACTION: ADD\nPATH: dir/b.txt\n---\nB\n---\n"
	res, err := ApplyPatch(patch)
	if err != nil || res.IsError { t.Fatalf("multiple add failed: %v %v", err, res.Output) }
	if _, err := os.Stat(filepath.Join(d, "a.txt")); err != nil { t.Fatalf("a.txt missing: %v", err) }
	if _, err := os.Stat(filepath.Join(d, "dir", "b.txt")); err != nil { t.Fatalf("dir/b.txt missing: %v", err) }
}

func TestApplyPatchPathTraversalAndAbsolute(t *testing.T) {
	d := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(d); err != nil { t.Fatalf("chdir: %v", err) }

	// path traversal
	patch1 := "ACTION: ADD\nPATH: ../evil.txt\n---\nX\n---\n"
	if _, err := ApplyPatch(patch1); err == nil {
		t.Fatalf("expected error for path traversal")
	}

	// absolute path rejection (platform aware)
	absPath := "/etc/passwd"
	if runtime.GOOS == "windows" {
		absPath = "C:\\Windows\\system32\\drivers\\etc\\hosts"
	}
	patch2 := "ACTION: ADD\nPATH: " + absPath + "\n---\nX\n---\n"
	if _, err := ApplyPatch(patch2); err == nil {
		t.Fatalf("expected error for absolute path")
	}
}

func TestApplyPatchAddExistingAndDeleteMissing(t *testing.T) {
	d := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(d); err != nil { t.Fatalf("chdir: %v", err) }

	// precreate file
	if err := os.WriteFile("exists.txt", []byte("old\n"), 0644); err != nil { t.Fatalf("precreate: %v", err) }
	patchAdd := "ACTION: ADD\nPATH: exists.txt\n---\nnew\n---\n"
	res, err := ApplyPatch(patchAdd)
	if err != nil || res.IsError {
		t.Fatalf("ADD on existing should skip: %v %v", err, res.Output)
	}

	// delete missing should skip
	patchDel := "ACTION: DELETE\nPATH: missing.txt\n---\n---\n"
	res, err = ApplyPatch(patchDel)
	if err != nil || res.IsError {
		t.Fatalf("DELETE missing should not error: %v %v", err, res.Output)
	}
}

func TestApplyPatchConflictingActions(t *testing.T) {
	d := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(d); err != nil { t.Fatalf("chdir: %v", err) }

	// ADD then UPDATE in same patch
	patch := "ACTION: ADD\nPATH: mix.txt\n---\nfirst\n---\n\nACTION: UPDATE\nPATH: mix.txt\n---\nsecond\n---\n"
	res, err := ApplyPatch(patch)
	if err != nil || res.IsError { t.Fatalf("conflicting actions failed: %v %v", err, res.Output) }
	b, err := os.ReadFile("mix.txt")
	if err != nil { t.Fatalf("read mix.txt: %v", err) }
	if string(b) != "second" { t.Fatalf("unexpected mix.txt content: %q", string(b)) }
}
