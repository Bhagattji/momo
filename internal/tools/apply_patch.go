package tools

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// copyFileContents copies a file from src to dst (overwrites dst).
func copyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	// try to copy mode
	if fi, err := in.Stat(); err == nil {
		_ = out.Chmod(fi.Mode())
	}
	return nil
}

// apply_patch supports a simple action block format (preview + apply):
// ACTION: ADD|UPDATE|DELETE
// PATH: <relative path>
// ---
// <content>
// ---

func ApplyPatch(patch string) (ToolResult, error) {
	lines := strings.Split(patch, "\n")
	if len(lines) == 0 {
		return ToolResult{IsError: true, Output: "empty patch"}, errors.New("empty patch")
	}

	type act struct{ Action, Path, Content string }
	actions := []act{}

	// parse simple single-action or multiple concatenated actions separated by blank lines
	i := 0
	for i < len(lines) {
		// skip empty lines
		if strings.TrimSpace(lines[i]) == "" { i++; continue }
		if !strings.HasPrefix(lines[i], "ACTION:") {
			return ToolResult{IsError: true, Output: "patch missing ACTION"}, errors.New("patch missing ACTION")
		}
		action := strings.TrimSpace(strings.TrimPrefix(lines[i], "ACTION:"))
		i++
		if i >= len(lines) || !strings.HasPrefix(lines[i], "PATH:") {
			return ToolResult{IsError: true, Output: "patch missing PATH"}, errors.New("patch missing PATH")
		}
		p := strings.TrimSpace(strings.TrimPrefix(lines[i], "PATH:"))
		i++
		// expect --- then content until ---
		if i < len(lines) && strings.TrimSpace(lines[i]) == "---" {
			i++
			start := i
			for i < len(lines) && strings.TrimSpace(lines[i]) != "---" { i++ }
			content := strings.Join(lines[start:i], "\n")
			// skip closing --- if present
			if i < len(lines) && strings.TrimSpace(lines[i]) == "---" { i++ }
			actions = append(actions, act{Action: action, Path: p, Content: content})
		} else {
			// no content block
			actions = append(actions, act{Action: action, Path: p, Content: ""})
		}
	}

	// Security checks & prepare transactional application
	cwd, err := os.Getwd()
	if err != nil { return ToolResult{IsError: true, Output: err.Error()}, err }

	// normalize cwd with trailing separator
	cwdAbs, _ := filepath.Abs(cwd)
	if !strings.HasSuffix(cwdAbs, string(os.PathSeparator)) {
		cwdAbs = cwdAbs + string(os.PathSeparator)
	}

	// Validate paths
	for _, a := range actions {
		p := filepath.Clean(a.Path)
		if filepath.IsAbs(p) {
			return ToolResult{IsError: true, Output: "invalid path (absolute): " + a.Path}, errors.New("invalid path")
		}
		// prevent traversal after cleaning
		if strings.HasPrefix(p, "..") || strings.Contains(p, ".."+string(os.PathSeparator)) {
			return ToolResult{IsError: true, Output: "invalid path (traversal): " + a.Path}, errors.New("invalid path")
		}
		// ensure resolved dir stays under cwd
		dir := filepath.Dir(p)
		if dir == "." { dir = "" }
		absDir := cwdAbs
		if dir != "" {
			absDirTmp, err := filepath.Abs(filepath.Join(cwd, dir))
			if err != nil { return ToolResult{IsError: true, Output: err.Error()}, err }
			absDir = absDirTmp
			if !strings.HasPrefix(absDir+string(os.PathSeparator), cwdAbs) {
				return ToolResult{IsError: true, Output: "invalid path (outside workspace): " + a.Path}, errors.New("invalid path")
			}
		}
	}

	// Apply actions transactionally: track backups and created files
	type backupEntry struct{ Path string; Backup string; WasExisting bool }
	backups := []backupEntry{}
	created := []string{}
	results := []string{}

	rollback := func() {
		// restore backups
		for i := len(backups)-1; i >= 0; i-- {
			b := backups[i]
			if b.WasExisting {
				// restore by copying backup back to path
				_ = copyFileContents(b.Backup, b.Path)
				_ = os.Remove(b.Backup)
			} else {
				// if it didn't exist before, remove target if present
				_ = os.Remove(b.Path)
			}
		}
		// remove created files
		for _, c := range created {
			_ = os.Remove(c)
		}
	}

	for _, a := range actions {
		p := filepath.Clean(a.Path)
		switch strings.ToUpper(a.Action) {
		case "ADD":
			if _, err := os.Stat(p); err == nil {
				results = append(results, fmt.Sprintf("skipped add: %s exists", p))
				continue
			}
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				rollback(); return ToolResult{IsError: true, Output: err.Error()}, err
			}
			// write to temp file in same dir then rename
			tmpf, err := os.CreateTemp(filepath.Dir(p), ".momo.patch.*")
			if err != nil { rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			if _, err := io.WriteString(tmpf, a.Content); err != nil { tmpf.Close(); os.Remove(tmpf.Name()); rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			tmpf.Close()
			if err := os.Rename(tmpf.Name(), p); err != nil { os.Remove(tmpf.Name()); rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			created = append(created, p)
			results = append(results, fmt.Sprintf("added %s", p))

		case "UPDATE":
			if _, err := os.Stat(p); os.IsNotExist(err) {
				rollback(); return ToolResult{IsError: true, Output: "update failed: file not found: " + p}, errors.New("file not found")
			}
			// create backup by copying original to temp (avoid cross-device rename issues)
			bakf, err := os.CreateTemp("", "momo.patch.bak.*")
			if err != nil { rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			bakPath := bakf.Name()
			bakf.Close()
			// copy contents
			if err := copyFileContents(p, bakPath); err != nil { os.Remove(bakPath); rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			backups = append(backups, backupEntry{Path: p, Backup: bakPath, WasExisting: true})
			// write new content to temp file then rename into place
			tmpf, err := os.CreateTemp(filepath.Dir(p), ".momo.patch.*")
			if err != nil { rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			if _, err := io.WriteString(tmpf, a.Content); err != nil { tmpf.Close(); os.Remove(tmpf.Name()); rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			tmpf.Close()
			if err := os.Rename(tmpf.Name(), p); err != nil {
				// attempt restore from backup copy
				_ = copyFileContents(bakPath, p)
				os.Remove(tmpf.Name())
				rollback()
				return ToolResult{IsError: true, Output: err.Error()}, err
			}
			results = append(results, fmt.Sprintf("updated %s", p))

		case "DELETE":
			if _, err := os.Stat(p); os.IsNotExist(err) {
				results = append(results, fmt.Sprintf("skipped delete: %s missing", p))
				continue
			}
			// backup before delete by copying
			bakf, err := os.CreateTemp("", "momo.patch.bak.*")
			if err != nil { rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			bakPath := bakf.Name()
			bakf.Close()
			if err := copyFileContents(p, bakPath); err != nil { os.Remove(bakPath); rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			// remove original
			if err := os.Remove(p); err != nil { os.Remove(bakPath); rollback(); return ToolResult{IsError: true, Output: err.Error()}, err }
			backups = append(backups, backupEntry{Path: p, Backup: bakPath, WasExisting: true})
			results = append(results, fmt.Sprintf("deleted %s", p))

		default:
			rollback(); return ToolResult{IsError: true, Output: "unknown action: " + a.Action}, errors.New("unknown action")
		}
	}

	return ToolResult{Title: "apply_patch", Output: strings.Join(results, "\n"), IsError: false}, nil
}
