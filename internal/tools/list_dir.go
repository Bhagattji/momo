package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ListDir(path string) (ToolResult, error) {
	if path == "" {
		path = "."
	}
	p := filepath.Clean(path)
	entries, err := os.ReadDir(p)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, err
	}
	var sb strings.Builder
	for _, e := range entries {
		info, _ := e.Info()
		if info.IsDir() {
			sb.WriteString("[DIR] " + e.Name() + "\n")
		} else {
			sb.WriteString(e.Name() + " (" + formatSize(info.Size()) + ")\n")
		}
	}
	return ToolResult{Title: "list_dir: " + p, Output: sb.String()}, nil
}

func formatSize(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%dB", n)
	}
	k := float64(n) / 1024.0
	if k < 1024 {
		return fmt.Sprintf("%.1fKB", k)
	}
	m := k / 1024.0
	return fmt.Sprintf("%.1fMB", m)
}
