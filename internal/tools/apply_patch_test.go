package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyPatchAddUpdateDelete(t *testing.T) {
	// work inside temp dir
	d := t.TempDir()
	prev, _ := os.Getwd()
	defer os.Chdir(prev)
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	// ADD
	patchAdd := "ACTION: ADD\nPATH: hello.txt\n---\nHello World\n---\n"
	res, err := ApplyPatch(patchAdd)
	if err != nil || res.IsError {
		t.Fatalf("ApplyPatch ADD failed: %v, %v", err, res.Output)
	}
	p := filepath.Join(d, "hello.txt")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read added file: %v", err)
	}
	if string(b) != "Hello World" {
		t.Fatalf("unexpected content after add: %q", string(b))
	}


	// UPDATE
	patchUpd := "ACTION: UPDATE\nPATH: hello.txt\n---\nUpdated\n---\n"
	res, err = ApplyPatch(patchUpd)
	if err != nil || res.IsError {
		t.Fatalf("ApplyPatch UPDATE failed: %v, %v", err, res.Output)
	}
	b, err = os.ReadFile(p)
	if err != nil {
		t.Fatalf("read updated file: %v", err)
	}
	if string(b) != "Updated" {
		t.Fatalf("unexpected content after update: %q", string(b))
	}


	// DELETE
	patchDel := "ACTION: DELETE\nPATH: hello.txt\n---\n---\n"
	res, err = ApplyPatch(patchDel)
	if err != nil || res.IsError {
		t.Fatalf("ApplyPatch DELETE failed: %v, %v", err, res.Output)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatalf("file not deleted: stat err=%v", err)
	}
}
