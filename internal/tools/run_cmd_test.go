package tools

import (
	"strings"
	"testing"
)

func TestRunCmdEcho(t *testing.T) {
	res, err := RunCmd("echo hello", "", 5)
	if err != nil && res.IsError {
		t.Fatalf("RunCmd failed: %v, output: %s", err, res.Output)
	}
	if !strings.Contains(res.Output, "hello") {
		t.Fatalf("expected hello in output, got: %q", res.Output)
	}
}
