package provider

import (
	"context"
	"errors"
)

// Build returns a Provider implementation for the given name.
// For now this returns an OpenAI-compatible stub for any name.
func Build(name, model, key, baseURL string) (Provider, string, error) {
	if name == "" {
		return nil, "", errors.New("provider name required")
	}
	p := &OpenAICompat{name: name, model: model, apiKey: key, baseURL: baseURL}
	if p.model == "" {
		p.model = "gpt-4o"
	}
	// quick validation: ListModels should not fail for stub
	_, err := p.ListModels(context.Background())
	if err != nil {
		return nil, "", err
	}
	return p, p.model, nil
}
