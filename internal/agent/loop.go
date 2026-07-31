package agent

import (
	"context"
	"momo/internal/provider"
)

type Step struct {
	Content   string
	ToolCalls []provider.ToolCall
	Usage     provider.Usage
}

type Budget struct {
	TotalInputTokens  int
	TotalOutputTokens int
	MaxTokens         int
}

func NewBudget(maxTokens int) *Budget {
	return &Budget{MaxTokens: maxTokens}
}

func (b *Budget) Track(usage provider.Usage) bool {
	b.TotalInputTokens += usage.InputTokens
	b.TotalOutputTokens += usage.OutputTokens
	total := b.TotalInputTokens + b.TotalOutputTokens
	if b.MaxTokens > 0 && total >= b.MaxTokens {
		return false
	}
	return true
}

func (b *Budget) Used() int {
	return b.TotalInputTokens + b.TotalOutputTokens
}

func Run(ctx context.Context, prov provider.Provider, exec *Executor, systemPrompt string, maxSteps int, onStep func(Step) error) error {
	return RunWithBudget(ctx, prov, exec, systemPrompt, maxSteps, nil, onStep)
}

func RunWithBudget(ctx context.Context, prov provider.Provider, exec *Executor, systemPrompt string, maxSteps int, budget *Budget, onStep func(Step) error) error {
	messages := []provider.Message{{Role: "system", Content: systemPrompt}}

	for stepNum := 0; stepNum < maxSteps; stepNum++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req := provider.CompletionRequest{
			Model:    prov.Model(),
			Messages: messages,
			Stream:   false,
		}
		resp, err := prov.Chat(ctx, req)
		if err != nil {
			onStep(Step{Content: "[error] " + err.Error()})
			return err
		}

		if budget != nil && !budget.Track(resp.Usage) {
			onStep(Step{Content: "[budget] token budget exceeded"})
			return nil
		}

		messages = append(messages, provider.Message{Role: "assistant", Content: resp.Content})

		s := Step{Content: resp.Content, ToolCalls: resp.ToolCalls, Usage: resp.Usage}
		if err := onStep(s); err != nil {
			return err
		}

		if len(resp.ToolCalls) == 0 {
			break
		}

		for _, tc := range resp.ToolCalls {
			res, err := exec.Execute(tc.Name, tc.Arguments)
			var output string
			if err != nil {
				output = "[executor error] " + err.Error()
			} else {
				output = res.Output
			}

			if res.IsError {
				output = "[tool-error] " + output
			}

			messages = append(messages, provider.Message{Role: "tool", Content: output})

			if err := onStep(Step{Content: output}); err != nil {
				return err
			}
		}
	}
	return nil
}