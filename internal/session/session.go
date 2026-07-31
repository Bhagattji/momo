package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"momo/internal/provider"
)

type Message struct {
	Role      string              `json:"role"`
	Content   string              `json:"content,omitempty"`
	ToolCalls []provider.ToolCall `json:"tool_calls,omitempty"`
	Timestamp string              `json:"timestamp"`
}

type Session struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Model      string    `json:"model"`
	Messages  []Message  `json:"messages"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

func New(providerName, modelName string) *Session {
	now := time.Now().UTC().Format(time.RFC3339)
	return &Session{
		ID:        fmt.Sprintf("ses_%x", time.Now().UnixNano()),
		Provider:  providerName,
		Model:      modelName,
		Messages:  []Message{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (s *Session) AddMessage(role, content string, toolCalls []provider.ToolCall) {
	s.Messages = append(s.Messages, Message{
		Role:      role,
		Content:   content,
		ToolCalls: toolCalls,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
}

func (s *Session) ToProviderMessages() []provider.Message {
	out := make([]provider.Message, 0, len(s.Messages))
	for _, m := range s.Messages {
		out = append(out, provider.Message{
			Role:      m.Role,
			Content:   m.Content,
			ToolCalls: m.ToolCalls,
		})
	}
	return out
}

func sessionDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(dir, "momo", "sessions")
	if err := os.MkdirAll(p, 0o700); err != nil {
		return "", err
	}
	return p, nil
}

func Save(s *Session) error {
	d, err := sessionDir()
	if err != nil {
		return err
	}
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	path := filepath.Join(d, s.ID+".json")
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func Load(id string) (*Session, error) {
	d, err := sessionDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(d, id+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func List() ([]Session, error) {
	d, err := sessionDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(d)
	if err != nil {
		return []Session{}, nil
	}
	var sessions []Session
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(d, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var s Session
		if err := json.Unmarshal(b, &s); err != nil {
			continue
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func Delete(id string) error {
	d, err := sessionDir()
	if err != nil {
		return err
	}
	path := filepath.Join(d, id+".json")
	return os.Remove(path)
}