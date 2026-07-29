package tools

import (
	"os"
	"path/filepath"
)

func WriteFile(path string, content string) (ToolResult, error) {
	if path == "" {
		return ToolResult{}, nil
	}
	p := filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, err
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, err
	}
	return ToolResult{Title: "write_file: " + p, Output: "wrote bytes", IsError: false}, nil
}
