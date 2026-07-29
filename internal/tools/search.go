package tools

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func Search(pattern string, path string) (ToolResult, error) {
	if path == "" { path = "." }
	p := filepath.Clean(path)
	var isRegex bool
	var re *regexp.Regexp
	if strings.HasPrefix(pattern, "regex:") {
		isRegex = true
		re = regexp.MustCompile(pattern[len("regex:"):])
	}
	var sb strings.Builder
	count := 0
	_ = filepath.WalkDir(p, func(fp string, d os.DirEntry, err error) error {
		if err != nil { return nil }
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "__pycache__" { return filepath.SkipDir }
			return nil
		}
		f, err := os.Open(fp)
		if err != nil { return nil }
		defer f.Close()
		s := bufio.NewScanner(f)
		lineno := 0
		for s.Scan() {
			lineno++
			line := s.Text()
			match := false
			if isRegex {
				if re.MatchString(line) { match = true }
			} else {
				if strings.Contains(line, pattern) { match = true }
			}
			if match {
				sb.WriteString(fp + ":" + strconv.Itoa(lineno) + ":" + line + "\n")
				count++
				if count >= 200 { return nil }
			}
		}
		return nil
	})
	return ToolResult{Title: "search", Output: sb.String()}, nil
}
