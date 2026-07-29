package tools

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

func RunCmd(command string, cwd string, timeout int) (ToolResult, error) {
	if command == "" {
		return ToolResult{}, nil
	}
	ctx := context.Background()
	if timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, "cmd", "/C", command)
	if cwd != "" {
		cmd.Dir = cwd
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		out := stdout.String() + "\n" + stderr.String()
		if len(out) > 20*1024 { out = out[:20*1024] }
		return ToolResult{IsError: true, Output: out}, err
	}
	out := stdout.String()
	if len(out) > 20*1024 { out = out[:20*1024] }
	return ToolResult{Title: "run_cmd", Output: out, IsError: false}, nil
}
