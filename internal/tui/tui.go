package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Permission request/response structs used by executor -> tui.
type PermissionRequest struct {
	ID         string
	Tool       string
	Args       string
	Message    string
	SaveToProject bool // if true, remembering approval will be saved to project config
}

type PermissionResponse struct {
	ID       string
	Approved bool
	Remember bool
}

// Channels for communicating permission requests. Initialized when Start() runs.
var PermissionRequests chan PermissionRequest
var PermissionResponses chan PermissionResponse

// Model is the BubbleTea model that can show a permission queue/dialog.
type Model struct{
	queue []PermissionRequest
	showing bool
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		if m.showing && len(m.queue) > 0 {
			switch strings.ToLower(v.String()) {
			case "y", "enter":
				req := m.queue[0]
				PermissionResponses <- PermissionResponse{ID: req.ID, Approved: true, Remember: false}
				// pop
				m.queue = m.queue[1:]
				if len(m.queue) == 0 { m.showing = false }
			case "r":
				req := m.queue[0]
				PermissionResponses <- PermissionResponse{ID: req.ID, Approved: true, Remember: true}
				m.queue = m.queue[1:]
				if len(m.queue) == 0 { m.showing = false }
			case "n", "esc":
				req := m.queue[0]
				PermissionResponses <- PermissionResponse{ID: req.ID, Approved: false, Remember: false}
				m.queue = m.queue[1:]
				if len(m.queue) == 0 { m.showing = false }
			}
		}
	case PermissionRequest:
		// append to queue and show dialog
		r := v
		m.queue = append(m.queue, r)
		m.showing = true
	}
	return m, nil
}

func (m Model) View() string {
	if m.showing && len(m.queue) > 0 {
		req := m.queue[0]
		args := req.Args
		if len(args) > 200 { args = args[:200] + "..." }
		saveHint := ""
		if req.SaveToProject {
			saveHint = " (will be saved to project .momo/config.json)"
		} else {
			saveHint = " (will be saved to global config)"
		}
		return fmt.Sprintf("Permission required (%d queued)\n\nTool: %s\nMessage: %s\nArgs: %s\n\nApprove (y/enter)%s  Approve & remember (r)%s  Reject (n/esc)", len(m.queue), req.Tool, req.Message, args, saveHint, saveHint)
	}
	return "momo TUI — placeholder\n\nPress Ctrl+C to quit."
}

// Start runs the TUI program in the background and initializes permission channels.
// Returns nil once the UI program is started; errors are logged to stderr.
func Start() error {
	PermissionRequests = make(chan PermissionRequest, 8)
	PermissionResponses = make(chan PermissionResponse, 8)

	p := tea.NewProgram(Model{})
	// Bridge PermissionRequests into the program's message loop
	go func() {
		for req := range PermissionRequests {
			p.Send(req)
		}
	}()

	// Run the program (blocking). The bridge goroutine keeps feeding requests.
	if err := p.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to start TUI:", err)
		return err
	}
	// Keep a short delay to let goroutine start
	time.Sleep(50 * time.Millisecond)
	return nil
}
