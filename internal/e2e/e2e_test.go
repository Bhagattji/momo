package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"momo/internal/provider"
)

func TestProviderConnectivity(t *testing.T) {
	key := os.Getenv("PROVIDER_API_KEY")
	if key == "" {
		t.Skip("PROVIDER_API_KEY not set — skipping provider connectivity test")
	}
	prov, _, err := provider.Build("openai", "", key, "")
	if err != nil {
		t.Fatalf("failed to build provider: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = prov.ListModels(ctx)
	if err != nil {
		t.Fatalf("provider ListModels failed: %v", err)
	}
}
