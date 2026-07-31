package provider

import "context"

// Provider is the interface for all LLM providers.
type Provider interface {
	Name() string
	Model() string
	SetModel(m string)
	Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Stream(ctx context.Context, req CompletionRequest, onChunk func(Chunk) error) error
	ListModels(ctx context.Context) ([]Model, error)
}

// Model represents an available model on a provider.
type Model struct {
	ID    string
	Owner string
}

// Message represents a chat message.
type Message struct {
	Role       string     `json:"role"`              // "user", "assistant", "tool", "system"
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"` // assistant → model requests tools
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// ToolCall represents a request from the model to call a tool.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // raw JSON string
}

// ToolDef describes a tool the model may call.
type ToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"` // JSON Schema
}

// CompletionRequest is the provider request payload.
type CompletionRequest struct {
	Model    string
	Messages []Message
	Tools    []ToolDef
	Stream   bool
	System   string
}

// CompletionResponse is the provider response.
type CompletionResponse struct {
	Content   string
	ToolCalls []ToolCall
	Usage     Usage
}

// Chunk represents a streaming chunk.
type Chunk struct {
	Content   string
	ToolCalls []ToolCall
	Done      bool
}

// Usage holds token counts.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}
