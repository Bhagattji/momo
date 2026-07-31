package tools

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func WriteFile(path string, content string) (ToolResult, error) {
	if path == "" {
		return ToolResult{}, nil
	}
	if err := validatePath(path); err != nil {
		return ToolResult{IsError: true, Output: "invalid path: " + err.Error()}, err
	}
	p := filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return ToolResult{IsError: true, Output: "failed to create directory: " + err.Error()}, err
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		return ToolResult{IsError: true, Output: "failed to write file: " + err.Error()}, err
	}
	return ToolResult{Title: "write_file: " + p, Output: "ok", IsError: false}, nil
}

func validatePath(path string) error {
	p := filepath.Clean(path)
	if filepath.IsAbs(p) && !isSubdirSafe(p) {
		return errors.New("absolute paths not allowed")
	}
	if strings.HasPrefix(p, "..") || strings.Contains(p, ".."+string(os.PathSeparator)) {
		return errors.New("path traversal not allowed")
	}
	return nil
}

func isSubdirSafe(p string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	cwdAbs, _ := filepath.Abs(cwd)
	if !strings.HasSuffix(cwdAbs, string(os.PathSeparator)) {
		cwdAbs = cwdAbs + string(os.PathSeparator)
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return false
	}
	return strings.HasPrefix(abs, cwdAbs)
}