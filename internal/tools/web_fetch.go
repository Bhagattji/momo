package tools

import (
	"io"
	"net/http"
	"strings"
	"time"
)

func WebFetch(url string) (ToolResult, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, err
	}
	defer resp.Body.Close()
	ct := resp.Header.Get("Content-Type")
	if !(strings.Contains(ct, "text/") || strings.Contains(ct, "json")) {
		return ToolResult{IsError: true, Output: "unsupported content-type: " + ct}, nil
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 30*1024))
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, err
	}
	s := string(b)
	// crude strip HTML tags
	s = stripTags(s)
	return ToolResult{Title: "web_fetch: " + url, Output: s}, nil
}

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' { inTag = true; continue }
		if r == '>' { inTag = false; continue }
		if !inTag { b.WriteRune(r) }
	}
	return b.String()
}
