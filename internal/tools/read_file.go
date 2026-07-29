package tools

import (
	"os"
	"path/filepath"
)

func ReadFile(path string, offset int, limit int) (ToolResult, error) {
	if path == "" {
		return ToolResult{}, nil
	}
	p := filepath.Clean(path)
	b, err := os.ReadFile(p)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, err
	}
	content := string(b)
	if offset > 0 && offset < len(content) {
		content = content[offset:]
	}
	if limit > 0 && len(content) > limit {
		content = content[:limit]
	}
	// truncate max 50KB
	if len(content) > 50*1024 {
		content = content[:50*1024]
	}
	return ToolResult{Title: "read_file: " + p, Output: content, IsError: false}, nil
}
