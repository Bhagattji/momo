package provider

// AuthKind enumerates how a provider authenticates.
type AuthKind string

const (
	AuthAPIKey AuthKind = "api-key"
	AuthLocal  AuthKind = "local" // no key needed (ollama)
)

// Info describes a known provider.
type Info struct {
	Name         string
	Env          string   // env var that holds the API key
	DefaultBase  string   // base URL (no trailing slash expected by openai_compat)
	DefaultModel string
	Notes        string
	Auth         AuthKind
	Free         bool // has a free tier
}

// Catalog returns the full list of known providers, sorted with free/fast first.
func Catalog() []Info {
	return []Info{
		// Tier 1 — free + fast
		{Name: "groq", Env: "GROQ_API_KEY", DefaultBase: "https://api.groq.com/openai", DefaultModel: "llama-3.3-70b-versatile", Notes: "Ultra-fast (free tier)", Auth: AuthAPIKey, Free: true},
		{Name: "cerebras", Env: "CEREBRAS_API_KEY", DefaultBase: "https://api.cerebras.ai/v1", DefaultModel: "llama3.1-70b", Notes: "Ultra-fast silicon (free)", Auth: AuthAPIKey, Free: true},
		{Name: "google", Env: "GOOGLE_API_KEY", DefaultBase: "https://generativelanguage.googleapis.com/v1beta", DefaultModel: "gemini-2.0-flash", Notes: "Gemini Flash (free)", Auth: AuthAPIKey, Free: true},

		// Tier 2 — free aggregators
		{Name: "openrouter", Env: "OPENROUTER_API_KEY", DefaultBase: "https://openrouter.ai/api/v1", DefaultModel: "meta-llama/llama-3.3-70b-instruct:free", Notes: "100+ models (free ones available)", Auth: AuthAPIKey, Free: true},
		{Name: "chutes", Env: "CHUTES_API_KEY", DefaultBase: "https://api.chutes.ai/v1", DefaultModel: "deepseek-ai/DeepSeek-V3", Notes: "Free tier", Auth: AuthAPIKey, Free: true},

		// Tier 3 — paid, powerful
		{Name: "anthropic", Env: "ANTHROPIC_API_KEY", DefaultBase: "https://api.anthropic.com", DefaultModel: "claude-3-5-sonnet-latest", Notes: "Claude 3.5 — best tools", Auth: AuthAPIKey},
		{Name: "openai", Env: "OPENAI_API_KEY", DefaultBase: "https://api.openai.com", DefaultModel: "gpt-4o-mini", Notes: "GPT-4o, o1, o3", Auth: AuthAPIKey},
		{Name: "deepseek", Env: "DEEPSEEK_API_KEY", DefaultBase: "https://api.deepseek.com/v1", DefaultModel: "deepseek-chat", Notes: "DeepSeek-V3 / R1", Auth: AuthAPIKey},
		{Name: "xai", Env: "XAI_API_KEY", DefaultBase: "https://api.x.ai/v1", DefaultModel: "grok-2-latest", Notes: "Grok family", Auth: AuthAPIKey},
		{Name: "mistral", Env: "MISTRAL_API_KEY", DefaultBase: "https://api.mistral.ai/v1", DefaultModel: "mistral-large-latest", Notes: "Mistral Large", Auth: AuthAPIKey},
		{Name: "nvidia", Env: "NVIDIA_API_KEY", DefaultBase: "https://integrate.api.nvidia.com/v1", DefaultModel: "meta/llama-3.1-70b-instruct", Notes: "NVIDIA NIM", Auth: AuthAPIKey},

		// Local
		{Name: "ollama", Env: "", DefaultBase: "http://localhost:11434/v1", DefaultModel: "llama3.1", Notes: "Local models (no key)", Auth: AuthLocal},
	}
}

// InfoByName looks up a provider by name. Returns nil if unknown.
func InfoByName(name string) *Info {
	for _, p := range Catalog() {
		if p.Name == name {
			cp := p
			return &cp
		}
	}
	return nil
}

// ResolveKey finds an API key for the provider, checking env vars.
// Order: provider's own env var → NAME_API_KEY → Momo's configured key.
func ResolveKey(name string, getEnv func(string) string, cfgKey string) string {
	if info := InfoByName(name); info != nil && info.Env != "" {
		if v := getEnv(info.Env); v != "" {
			return v
		}
	}
	if v := getEnv(name + "_API_KEY"); v != "" {
		return v
	}
	return cfgKey
}
