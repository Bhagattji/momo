package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
	modelFlg := flag.String("model", "", "Model to use (overrides config)")
	promptFlag := flag.String("prompt", "", "Send a one-shot prompt and exit")
	auto := flag.Bool("auto", false, "Auto-approve tool executions")
	debug := flag.Bool("debug", false, "Enable debug logging")
	showVersion := flag.Bool("version", false, "Print version and exit")
	timeoutSecs := flag.Int("timeout", 300, "Agent timeout in seconds")
	maxSteps := flag.Int("steps", 5, "Max agent steps per turn")
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

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	_ = config.Validate(cfg)

	effectiveProvider := cfg.DefaultProvider
	if *providerFlag != "" {
		effectiveProvider = *providerFlag
	}
	effectiveModel := cfg.DefaultModel
	if *modelFlg != "" {
		effectiveModel = *modelFlg
	}

	apiKey := resolveProviderKey(effectiveProvider, cfg.ProviderKey)
	baseURL := cfg.ProviderBaseURL

	logging.Info("Starting momo provider=%s model=%s auto=%v", effectiveProvider, effectiveModel, *auto)

	go func() {
		if err := tui.Start(); err != nil {
			logging.Warn("TUI not available: %v", err)
		}
	}()

	prov, modelUsed, err := provider.Build(effectiveProvider, effectiveModel, apiKey, baseURL)
	if err != nil {
		log.Fatalf("failed to build provider: %v", err)
	}
	logging.Info("Provider ready: %s / %s", effectiveProvider, modelUsed)

	r := tools.NewRegistry()
	exec := &agent.Executor{Registry: r, Auto: *auto, Workspace: cfg.Workspace, Cfg: cfg}

	if *timeoutSecs < 10 {
		*timeoutSecs = 300
	}

	if *promptFlag != "" {
		runSinglePrompt(*promptFlag, prov, exec, *timeoutSecs, *maxSteps)
		return
	}

	runInteractive(prov, exec, *timeoutSecs, *maxSteps)
}

func runSinglePrompt(prompt string, prov provider.Provider, exec *agent.Executor, timeoutSecs int, maxSteps int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	_ = agent.Run(ctx, prov, exec, prompt, maxSteps, func(s agent.Step) error {
		fmt.Println(s.Content)
		return nil
	})
}

func runInteractive(prov provider.Provider, exec *agent.Executor, timeoutSecs, maxSteps int) {
	history := []provider.Message{{Role: "system", Content: "You are Momo, a coding assistant. Be concise, respond in plain text, use tools when needed."}}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		switch input {
		case "/exit", "/quit", "/q":
			fmt.Println("Goodbye.")
			return
		case "/help", "/?":
			fmt.Println("/exit, /quit, /q - exit")
			fmt.Println("/help - show this help")
			continue
		}

		history = append(history, provider.Message{Role: "user", Content: input})

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)

		req := provider.CompletionRequest{
			Model:    prov.Model(),
			Messages: history,
			Stream:   false,
		}

		resp, err := prov.Chat(ctx, req)
		if err != nil {
			fmt.Println("[error]", err)
			cancel()
			continue
		}

		history = append(history, provider.Message{Role: "assistant", Content: resp.Content})
		fmt.Println(resp.Content)

		for _, tc := range resp.ToolCalls {
			output := fmt.Sprintf("> %s(%s)", tc.Name, tc.Arguments)
			fmt.Println(output)
			res, execErr := exec.Execute(tc.Name, tc.Arguments)
			if execErr != nil {
				fmt.Println("[tool error]", execErr)
			}
			if res.Output != "" {
				fmt.Println(res.Output)
			}
		}

		cancel()
	}
	fmt.Println()
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