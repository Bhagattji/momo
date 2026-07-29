package agent

import (
	"fmt"
	"time"

	"momo/internal/config"
	"momo/internal/tools"
	"momo/internal/tui"
)

// Executor runs tools with permission checks.
type Executor struct {
	Registry *tools.Registry
	Auto     bool
	Workspace string
	Cfg      *config.Config
}

// Execute runs a tool by name with given args JSON.
// Write/exec tools require approval unless Executor.Auto is true.
func (e *Executor) Execute(name string, argsJSON string) (tools.ToolResult, error) {
	// If no registry, nothing to do.
	if e.Registry == nil {
		return tools.ToolResult{IsError: true, Output: "no tool registry available"}, nil
	}

	// Define risk map: read is safe, write/exec are risky.
	risky := map[string]bool{
		"write_file": true,
		"edit_file":  true,
		"run_cmd":    true,
		"apply_patch": true,
	}

	// If config says this tool is approved, skip prompt.
	if e.Cfg != nil {
		if approved, ok := e.Cfg.Approvals[name]; ok && approved {
			return e.Registry.Execute(name, argsJSON)
		}
	}

	if risky[name] && !e.Auto {
		// Send permission request to TUI and wait for response (timeout 60s).
		reqID := fmt.Sprintf("%d", time.Now().UnixNano())
		req := tui.PermissionRequest{ID: reqID, Tool: name, Args: argsJSON, Message: fmt.Sprintf("Approve tool %s?", name), SaveToProject: e.Workspace != ""}
		// send request
		select {
		case tui.PermissionRequests <- req:
			// sent
		default:
			// if TUI not running or channel blocked, fallback to permission-needed result
			msg := fmt.Sprintf("[permission required] tool '%s' requires approval but TUI unavailable. Run with --auto.", name)
			return tools.ToolResult{IsError: true, Output: msg}, nil
		}

		// wait for response with timeout
		timeout := time.After(60 * time.Second)
		for {
			select {
			case resp := <-tui.PermissionResponses:
				if resp.ID != reqID {
					// not our response — ignore (other requests)
					continue
				}
				if !resp.Approved {
					return tools.ToolResult{IsError: true, Output: "permission denied"}, nil
				}
				// If user asked to remember approval, persist it in config.
				if resp.Remember {
				if e.Cfg != nil {
					if e.Cfg.Approvals == nil { e.Cfg.Approvals = map[string]bool{} }
					e.Cfg.Approvals[name] = true
					// persist to project config if workspace provided, else global
					if e.Workspace != "" {
						_ = config.SaveProject(e.Cfg, e.Workspace)
					} else {
						_ = config.Save(e.Cfg)
					}
				}
				}
				// execute
				res, err := e.Registry.Execute(name, argsJSON)
				if err != nil {
					return tools.ToolResult{IsError: true, Output: err.Error()}, nil
				}
				return res, nil
			case <-timeout:
				return tools.ToolResult{IsError: true, Output: "permission timeout"}, nil
			}
		}
	}

	// Non-risky or auto mode: execute via registry
	res, err := e.Registry.Execute(name, argsJSON)
	if err != nil {
		// Wrap registry error into ToolResult
		return tools.ToolResult{IsError: true, Output: err.Error()}, nil
	}
	return res, nil
}
