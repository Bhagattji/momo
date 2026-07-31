package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"momo/internal/agent"
	"momo/internal/config"
	"momo/internal/logging"
	"momo/internal/provider"
	"momo/internal/tools"
	"momo/internal/tui"
	"momo/internal/version"
)

func main() {
	providerFlag := flag.String("provider", "", "LLM provider to use (overrides config)")
	model := flag.String("model", "", "Model to use (overrides config)")
	auto := flag.Bool("auto", false, "Auto-approve tool executions")
	debug := flag.Bool("debug", false, "Enable debug logging")
	showVersion := flag.Bool("version", false, "Print version and exit")
	timeoutSecs := flag.Int("timeout", 120, "Agent timeout in seconds")
	flag.Parse()

	if len(flag.Args()) > 0 {
		sub := flag.Args()[0]
		subArgs := flag.Args()[1:]
		switch sub {
		case "self-update", "selfupdate":
			if err := selfUpdateCmd(subArgs); err != nil {
				log.Fatalf("self-update failed: %v", err)
			}
			return
		}
	}

	if *showVersion {
		fmt.Println("momo", version.Version)
		return
	}

	logging.Init(*debug)

	if *debug {
		logging.Debug("flags: provider=%s model=%s auto=%v timeout=%d", *providerFlag, *model, *auto, *timeoutSecs)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := config.Validate(cfg); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	effectiveProvider := cfg.DefaultProvider
	if *providerFlag != "" {
		effectiveProvider = *providerFlag
	}
	effectiveModel := cfg.DefaultModel
	if *model != "" {
		effectiveModel = *model
	}

	apiKey := resolveProviderKey(effectiveProvider, cfg.ProviderKey)
	baseURL := cfg.ProviderBaseURL

	fmt.Printf("Starting momo CLI\n  provider: %s\n  model:    %s\n  auto:     %v\n", effectiveProvider, effectiveModel, *auto)

	go func() {
		if err := tui.Start(); err != nil {
			log.Println("TUI error (continuing without TUI):", err)
		}
	}()

	prov, modelUsed, err := provider.Build(effectiveProvider, effectiveModel, apiKey, baseURL)
	if err != nil {
		log.Fatalf("failed to build provider: %v", err)
	}
	logging.Info("Using provider %s model %s", effectiveProvider, modelUsed)

	r := tools.NewRegistry()
	exec := &agent.Executor{Registry: r, Auto: *auto, Workspace: cfg.Workspace, Cfg: cfg}

	if *timeoutSecs < 10 {
		*timeoutSecs = 120
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSecs)*time.Second)
	defer cancel()

	systemPrompt := "You are Momo assistant. Be concise."
	if err := agent.Run(ctx, prov, exec, systemPrompt, 3, func(s agent.Step) error {
		fmt.Println("=== STEP ===")
		fmt.Println(s.Content)
		return nil
	}); err != nil {
		log.Println("agent run error:", err)
	}

	fmt.Println("momo run complete")
	logging.Shutdown()
}

func resolveProviderKey(name, cfgKey string) string {
	if v := provider.ResolveKey(name, os.Getenv, cfgKey); v != "" {
		return v
	}
	if v := os.Getenv("MOMO_API_KEY"); v != "" {
		return v
	}
	return cfgKey
}