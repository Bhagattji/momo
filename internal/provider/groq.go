package provider

import "context"

// Groq provider stub
type Groq struct {
	name   string
	model  string
	apiKey string
	baseURL string
}

func NewGroq(apiKey, baseURL, model string) *Groq {
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	if baseURL == "" {
		baseURL = "https://api.groq.com"
	}
	return &Groq{name: "groq", model: model, apiKey: apiKey, baseURL: baseURL}
}

func (g *Groq) Name() string { return g.name }
func (g *Groq) Model() string { return g.model }
func (g *Groq) SetModel(m string) { g.model = m }

func (g *Groq) Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// TODO: implement real HTTP call. Stub returns empty content for now.
	return &CompletionResponse{Content: ""}, nil
}

func (g *Groq) Stream(ctx context.Context, req CompletionRequest, onChunk func(Chunk) error) error {
	// No streaming in stub
	return nil
}

func (g *Groq) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{{ID: g.model, Owner: g.name}}, nil
}
