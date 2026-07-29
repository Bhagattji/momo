package provider

import "context"

// Anthropic provider stub
type Anthropic struct {
	name   string
	model  string
	apiKey string
	baseURL string
}

func NewAnthropic(apiKey, baseURL, model string) *Anthropic {
	if model == "" {
		model = "claude-3-5-sonnet-latest"
	}
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	return &Anthropic{name: "anthropic", model: model, apiKey: apiKey, baseURL: baseURL}
}

func (a *Anthropic) Name() string { return a.name }
func (a *Anthropic) Model() string { return a.model }
func (a *Anthropic) SetModel(m string) { a.model = m }

func (a *Anthropic) Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// TODO: implement real HTTP call
	return &CompletionResponse{Content: ""}, nil
}

func (a *Anthropic) Stream(ctx context.Context, req CompletionRequest, onChunk func(Chunk) error) error {
	return nil
}

func (a *Anthropic) ListModels(ctx context.Context) ([]Model, error) {
	return []Model{{ID: a.model, Owner: a.name}}, nil
}
