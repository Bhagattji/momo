package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PermissionRequest struct {
	ID            string
	Tool          string
	Args          string
	Message        string
	SaveToProject bool
}

type PermissionResponse struct {
	ID       string
	Approved bool
	Remember bool
}

type ChatMessage struct {
	Role    string
	Content string
}

var (
	PermissionRequests chan PermissionRequest
	PermissionResponses chan PermissionResponse

	internalReqs  = make(chan PermissionRequest, 16)
	internalResps = make(chan PermissionResponse, 16)
	internalChat  = make(chan ChatMessage, 64)

	started  bool
)

var (
	styleHighlight = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	styleAsst       = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	styleTool      = lipgloss.NewStyle().Foreground(lipgloss.Color("183"))
	styleErr       = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	styleDim       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	styleKey       = lipgloss.NewStyle().Foreground(lipgloss.Color("123")).Bold(true)
	stylePermHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
)

type model struct {
	chatHistory []ChatMessage
	permQueue   []PermissionRequest
	permActive  bool
	width       int
	height      int
	quit        bool
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		if m.permActive && len(m.permQueue) > 0 {
			switch strings.ToLower(v.String()) {
			case "y", "enter":
				r := m.permQueue[0]
				internalResps <- PermissionResponse{ID: r.ID, Approved: true, Remember: false}
				m.permQueue = m.permQueue[1:]
				if len(m.permQueue) == 0 {
					m.permActive = false
				}
			case "r":
				r := m.permQueue[0]
				internalResps <- PermissionResponse{ID: r.ID, Approved: true, Remember: true}
				m.permQueue = m.permQueue[1:]
				if len(m.permQueue) == 0 {
					m.permActive = false
				}
			case "n", "esc":
				r := m.permQueue[0]
				internalResps <- PermissionResponse{ID: r.ID, Approved: false, Remember: false}
				m.permQueue = m.permQueue[1:]
				if len(m.permQueue) == 0 {
					m.permActive = false
				}
			}
		} else {
			if v.String() == "ctrl+c" {
				m.quit = true
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height

	case PermissionRequest:
		m.permQueue = append(m.permQueue, v)
		m.permActive = true

	case ChatMessage:
		m.chatHistory = append(m.chatHistory, v)
	}
	return m, nil
}

func (m *model) View() string {
	if m.quit {
		return styleDim.Render("Goodbye.") + "\n"
	}
	if m.permActive && len(m.permQueue) > 0 {
		return m.permView()
	}
	return m.chatView()
}

func (m *model) chatView() string {
	var sb strings.Builder
	sb.WriteString(styleHighlight.Render("momo"))
	sb.WriteString(" ")
	sb.WriteString(styleDim.Render("AI Coding Agent | ctrl+c to quit"))
	sb.WriteString("\n\n")

	if len(m.chatHistory) == 0 {
		sb.WriteString(styleDim.Render("Ask me anything about your code..."))
		sb.WriteString("\n")
		return sb.String()
	}

	for _, msg := range m.chatHistory {
		switch msg.Role {
		case "assistant":
			for _, line := range strings.Split(msg.Content, "\n") {
				sb.WriteString(styleAsst.Render("  "))
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		case "tool_output":
			content := msg.Content
			if len(content) > 150 {
				content = content[:150] + "..."
			}
			sb.WriteString(styleTool.Render("  + "))
			sb.WriteString(content)
			sb.WriteString("\n")
		case "tool":
			sb.WriteString(styleTool.Render("  > "))
			sb.WriteString(msg.Content)
			sb.WriteString("\n")
		case "error":
			sb.WriteString(styleErr.Render("  ! "))
			sb.WriteString(msg.Content)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (m *model) permView() string {
	var sb strings.Builder
	req := m.permQueue[0]

	sb.WriteString(stylePermHeader.Render("=== PERMISSION REQUIRED ==="))
	sb.WriteString("\n")
	if len(m.permQueue) > 1 {
		sb.WriteString(fmt.Sprintf("  Pending: %d more in queue\n", len(m.permQueue)-1))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  Tool:  %s\n", req.Tool))
	if req.Args != "" {
		args := req.Args
		if len(args) > 120 {
			args = args[:120] + "..."
		}
		sb.WriteString(fmt.Sprintf("  Args:  %s\n", args))
	}
	sb.WriteString("\n")
	sb.WriteString(styleKey.Render("  y/enter"))
	sb.WriteString(" = Approve   ")
	sb.WriteString(styleKey.Render("r"))
	sb.WriteString(" = Approve+Remember   ")
	sb.WriteString(styleKey.Render("n/esc"))
	sb.WriteString(" = Deny")
	sb.WriteString("\n\n")

	if req.SaveToProject {
		sb.WriteString(styleDim.Render("  Remember saves to project .momo/config.json"))
	} else {
		sb.WriteString(styleDim.Render("  Remember saves to global config"))
	}
	return sb.String()
}

func init() {
	PermissionRequests = internalReqs
	PermissionResponses = internalResps
}

func Start() error {
	if started {
		return nil
	}
	started = true

	m := &model{}

	p := tea.NewProgram(m)

	go func() {
		for {
			select {
			case req := <-internalReqs:
				p.Send(req)
			case msg := <-internalChat:
				p.Send(msg)
			}
		}
	}()

	go func() {
		if err := p.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "tui error:", err)
		}
	}()

	return nil
}