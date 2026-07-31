package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DefaultProvider string          `json:"default_provider"`
	DefaultModel    string          `json:"default_model"`
	Approvals       map[string]bool `json:"approvals,omitempty"`
	ProviderKey     string          `json:"provider_key,omitempty"`
	ProviderBaseURL string          `json:"provider_base_url,omitempty"`
	Workspace       string          `json:"workspace,omitempty"`
}

func Load() (*Config, error) {
	cfg := &Config{DefaultProvider: "groq", DefaultModel: "llama-3.3-70b-versatile", Approvals: map[string]bool{}}

	dir, err := os.UserConfigDir()
	if err == nil {
		appDir := filepath.Join(dir, "momo")
		path := filepath.Join(appDir, "config.json")
		if b, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(b, cfg)
		}
	}
	if cfg.Approvals == nil {
		cfg.Approvals = map[string]bool{}
	}

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
					if pcfg.ProviderBaseURL != "" {
						cfg.ProviderBaseURL = pcfg.ProviderBaseURL
					}
					if pcfg.Workspace != "" {
						cfg.Workspace = pcfg.Workspace
					}
				}
				break
			}
			parent := filepath.Dir(p)
			if parent == p {
				break
			}
			p = parent
		}
	}

	cfg.ResolveEnv()

	return cfg, nil
}

func (c *Config) ResolveEnv() {
	if v := os.Getenv("MOMO_WORKSPACE"); v != "" {
		c.Workspace = v
	}
	if v := os.Getenv("MOMO_PROVIDER_BASE_URL"); v != "" {
		c.ProviderBaseURL = v
	}
	if v := os.Getenv("MOMO_API_KEY"); v != "" {
		c.ProviderKey = v
	}
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
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	_ = os.Chmod(path, 0o600)
	return nil
}

func SaveProject(cfg *Config, workspace string) error {
	if workspace == "" {
		return Save(cfg)
	}
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

func Validate(cfg *Config) error {
	if cfg.DefaultProvider == "" {
		cfg.DefaultProvider = "groq"
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "llama-3.3-70b-versatile"
	}
	if cfg.ProviderBaseURL != "" {
		if !hasValidScheme(cfg.ProviderBaseURL) {
			cfg.ProviderBaseURL = ""
		}
	}
	return nil
}

func hasValidScheme(url string) bool {
	return len(url) > 4 && (url[:4] == "http" || url[:5] == "https")
}