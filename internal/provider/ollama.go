package provider

import "context"

// Ollama provider stub (local)
type Ollama struct {
	name   string
	model  string
	baseURL string
}

func NewOllama(baseURL, model string) *Ollama {
	if model == "" {
		model = "llama3.1"
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Ollama{name: "ollama", model: model, baseURL: baseURL}
}

func (o *Ollama) Name() string { return o.name }
func (o *Ollama) Model() string { return o.model }
func (o *Ollama) SetModel(m string) { o.model = m }

func (o *Ollama) Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// TODO: call local Ollama API
	return &CompletionResponse{Content: ""}, nil
}

func (o *Ollama) Stream(ctx context.Context, req CompletionRequest, onChunk func(Chunk) error) error {
	return nil
}

func (o *Ollama) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{{ID: o.model, Owner: o.name}}, nil
}
