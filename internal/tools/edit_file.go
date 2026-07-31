package tools

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func EditFile(path string, oldString string, newString string) (ToolResult, error) {
	if path == "" {
		return ToolResult{}, nil
	}
	if err := validatePath(path); err != nil {
		return ToolResult{IsError: true, Output: "invalid path: " + err.Error()}, err
	}
	p := filepath.Clean(path)
	b, err := os.ReadFile(p)
	if err != nil {
		return ToolResult{IsError: true, Output: "failed to read file: " + err.Error()}, err
	}
	s := string(b)
	idx := strings.Index(s, oldString)
	if idx == -1 {
		return ToolResult{IsError: true, Output: "old_string not found in file"}, errors.New("old_string not found")
	}
	s = s[:idx] + newString + s[idx+len(oldString):]
	if err := os.WriteFile(p, []byte(s), 0o644); err != nil {
		return ToolResult{IsError: true, Output: "failed to write file: " + err.Error()}, err
	}
	return ToolResult{Title: "edit_file: " + p, Output: "edited (1 replacement)", IsError: false}, nil
}