package tools

import (
	"encoding/json"
	"fmt"
)

// Registry holds available tools and executes them.
type Registry struct{
}

func NewRegistry() *Registry { return &Registry{} }

// Execute dispatches to built-in tool implementations based on name.
func (r *Registry) Execute(name string, argsJSON string) (ToolResult, error) {
	switch name {
	case "read_file":
		var args struct{ Path string `json:"path"`; Offset int `json:"offset"`; Limit int `json:"limit"` }
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ToolResult{}, err
		}
		return ReadFile(args.Path, args.Offset, args.Limit)
	case "list_dir":
		var args struct{ Path string `json:"path"` }
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ToolResult{}, err
		}
		return ListDir(args.Path)
	case "run_cmd":
		var args struct{ Command string `json:"command"`; Cwd string `json:"cwd"`; Timeout int `json:"timeout"` }
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ToolResult{}, err
		}
		return RunCmd(args.Command, args.Cwd, args.Timeout)
	case "write_file":
		var args struct{ Path string `json:"path"`; Content string `json:"content"` }
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ToolResult{}, err
		}
		return WriteFile(args.Path, args.Content)
	case "edit_file":
		var args struct{ Path string `json:"path"`; OldString string `json:"old_string"`; NewString string `json:"new_string"` }
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ToolResult{}, err
		}
		return EditFile(args.Path, args.OldString, args.NewString)
	case "search":
		var args struct{ Pattern string `json:"pattern"`; Path string `json:"path"` }
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ToolResult{}, err
		}
		return Search(args.Pattern, args.Path)
	case "web_fetch":
		var args struct{ URL string `json:"url"` }
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ToolResult{}, err
		}
		return WebFetch(args.URL)
	case "apply_patch":
		var args struct{ Patch string `json:"patch"` }
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ToolResult{}, err
		}
		return ApplyPatch(args.Patch)
	default:
		return ToolResult{}, fmt.Errorf("unknown tool: %s", name)
	}
}
