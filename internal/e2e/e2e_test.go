package e2e

import (
	"os"
	"testing"

	"momo/internal/provider"
)

func TestProviderConnectivity(t *testing.T) {
	key := os.Getenv("PROVIDER_API_KEY")
	if key == "" {
		t.Skip("PROVIDER_API_KEY not set — skipping provider connectivity test")
	}
	// attempt to build provider (uses OpenAICompat stub). This checks basic HTTP calls.
	prov, _, err := provider.Build("openai", "", key, "")
	if err != nil {
		t.Fatalf("failed to build provider: %v", err)
	}
	// quick ListModels call with short timeout
	_, err = prov.ListModels(nil)
	if err != nil {
		t.Fatalf("provider ListModels failed: %v", err)
	}
}
