package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Minimal config for initial scaffold.
type Config struct {
	DefaultProvider string            `json:"default_provider"`
	DefaultModel    string            `json:"default_model"`
	Approvals       map[string]bool   `json:"approvals,omitempty"` // tool -> approved
	// Optional runtime fields (not recommended to store secrets here)
	ProviderKey     string            `json:"provider_key,omitempty"`
	ProviderBaseURL string            `json:"provider_base_url,omitempty"`
	Workspace       string            `json:"workspace,omitempty"`
}

func Load() (*Config, error) {
	// Load global config then merge project-level approvals if present.
	cfg := &Config{DefaultProvider: "groq", DefaultModel: "llama-3.3-70b-versatile", Approvals: map[string]bool{}}

	// Load global
	dir, err := os.UserConfigDir()
	if err == nil {
		appDir := filepath.Join(dir, "momo")
		path := filepath.Join(appDir, "config.json")
		if b, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(b, cfg)
		}
	}
	if cfg.Approvals == nil { cfg.Approvals = map[string]bool{} }

	// Merge project-level config (walk up from cwd)
	if wd, err := os.Getwd(); err == nil {
		p := wd
		for {
			projFile := filepath.Join(p, ".momo", "config.json")
			if b, err := os.ReadFile(projFile); err == nil {
				var pcfg Config
				if err := json.Unmarshal(b, &pcfg); err == nil {
					for k, v := range pcfg.Approvals {
						cfg.Approvals[k] = v
					}
					// merge non-sensitive runtime fields if present
					if pcfg.ProviderBaseURL != "" { cfg.ProviderBaseURL = pcfg.ProviderBaseURL }
					if pcfg.Workspace != "" { cfg.Workspace = pcfg.Workspace }
				}
				break
			}
			parent := filepath.Dir(p)
			if parent == p { break }
			p = parent
		}
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(dir, "momo")
	if err := os.MkdirAll(appDir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(appDir, "config.json")
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	// Write file with restrictive permissions where supported.
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	_ = os.Chmod(path, 0o600)
	return nil
}

// SaveProject persists a project-level config under workspace/.momo/config.json
func SaveProject(cfg *Config, workspace string) error {
	if workspace == "" { return Save(cfg) }
	projDir := filepath.Join(workspace, ".momo")
	if err := os.MkdirAll(projDir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(projDir, "config.json")
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	return nil
}
