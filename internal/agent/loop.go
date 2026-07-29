package agent

import (
	"context"
	"momo/internal/provider"
)

// Step represents an agent step emitted during Run.
type Step struct {
	Content   string
	ToolCalls []provider.ToolCall
	Usage     provider.Usage
}

// Run runs the agent loop. It maintains a message history, calls the provider,
// emits streaming/non-streaming steps via onStep, executes tool calls via exec,
// and appends tool results back into the conversation so the model can see them.
func Run(ctx context.Context, prov provider.Provider, exec *Executor, systemPrompt string, maxSteps int, onStep func(Step) error) error {
	// messages holds conversation history passed to provider on each turn.
	messages := []provider.Message{{Role: "system", Content: systemPrompt}}

	for stepNum := 0; stepNum < maxSteps; stepNum++ {
		req := provider.CompletionRequest{
			Model:    prov.Model(),
			Messages: messages,
			Stream:   false,
		}
		resp, err := prov.Chat(ctx, req)
		if err != nil {
			return err
		}

		// Append assistant response to history so model context grows.
		messages = append(messages, provider.Message{Role: "assistant", Content: resp.Content})

		// Emit the assistant content as a step for UI.
		s := Step{Content: resp.Content, ToolCalls: resp.ToolCalls, Usage: resp.Usage}
		if err := onStep(s); err != nil {
			return err
		}

		if len(resp.ToolCalls) == 0 {
			// No tools requested — agent finished.
			break
		}

		// Execute each tool call in order, append tool results to messages,
		// and emit the tool output as steps so UI can show progress.
		for _, tc := range resp.ToolCalls {
			res, err := exec.Execute(tc.Name, tc.Arguments)
			var output string
			if err != nil {
				output = "[executor error] " + err.Error()
			} else {
				output = res.Output
			}

			// If executor returned a permission-required style ToolResult with IsError,
			// surface that in output but still append to messages so model can react.
			if res.IsError {
				output = "[tool-error] " + output
			}

			// Append tool result as a tool message so the model sees results.
			messages = append(messages, provider.Message{Role: "tool", Content: output})

			// Emit tool result as a step for UI.
			if err := onStep(Step{Content: output}); err != nil {
				return err
			}
		}
	}
	return nil
}
