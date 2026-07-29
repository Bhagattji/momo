package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"momo/internal/agent"
	"momo/internal/config"
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
	flag.Parse()

		// Handle subcommands (e.g., momo self-update ...)
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

		if *debug {
			log.Println("debug: flags:", "provider=", *providerFlag, "model=", *model, "auto=", *auto)
		}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	effectiveProvider := cfg.DefaultProvider
		if *providerFlag != "" {
			effectiveProvider = *providerFlag
	}
	effectiveModel := cfg.DefaultModel
	if *model != "" {
		effectiveModel = *model
	}

		fmt.Printf("Starting momo CLI\n  provider: %s\n  model:    %s\n  auto:     %v\n", effectiveProvider, effectiveModel, *auto)

	// Start TUI in background so executor can use its channels.
	go func() {
		if err := tui.Start(); err != nil {
			log.Println("TUI error (continuing without TUI):", err)
		}
	}()

	// Build provider (uses OpenAICompat by default)
	prov, modelUsed, err := provider.Build(effectiveProvider, effectiveModel, "", "")
	if err != nil {
		log.Fatalf("failed to build provider: %v", err)
	}
	log.Printf("Using provider %s model %s", effectiveProvider, modelUsed)

	// Create tools registry and executor
	r := tools.NewRegistry()
		exec := &agent.Executor{Registry: r, Auto: *auto, Workspace: "", Cfg: cfg}

	// Run agent loop (simple example): short-lived demonstration
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	systemPrompt := "You are Momo assistant. Be concise."
	if err := agent.Run(ctx, prov, exec, systemPrompt, 3, func(s agent.Step) error {
		// For now print steps to stdout; TUI will surface permission dialogs.
		fmt.Println("=== STEP ===")
		fmt.Println(s.Content)
		return nil
	}); err != nil {
		log.Println("agent run error:", err)
	}

	fmt.Println("momo run complete")
}
