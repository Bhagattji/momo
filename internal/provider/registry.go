package provider

import (
	"errors"
)

// Build returns a Provider implementation for the given name.
// It resolves the base URL from the catalog and applies the key/model overrides.
func Build(name, model, key, baseURL string) (Provider, string, error) {
	if name == "" {
		return nil, "", errors.New("provider name required")
	}

	info := InfoByName(name)
	if info == nil {
		// Unknown provider → treat as generic OpenAI-compatible.
		info = &Info{Name: name, DefaultBase: baseURL, DefaultModel: model, Auth: AuthAPIKey}
	}

	if model == "" {
		model = info.DefaultModel
	}
	if baseURL == "" {
		baseURL = info.DefaultBase
	}

	// All current providers go through the OpenAI-compatible path.
	p := &OpenAICompat{name: name, model: model, apiKey: key, baseURL: baseURL}
	return p, model, nil
}
