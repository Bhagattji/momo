                                package agent

import (
	"context"
	"testing"
	"momo/internal/provider"
	"momo/internal/tools"
)

// fakeProv implements provider.Provider for tests.
type fakeProv struct{
	model string
	responses []*provider.CompletionResponse
	idx int
}

func (f *fakeProv) Name() string { return "fake" }
func (f *fakeProv) Model() string { return f.model }
func (f *fakeProv) SetModel(m string) { f.model = m }
func (f *fakeProv) Chat(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
	if f.idx >= len(f.responses) { return &provider.CompletionResponse{Content: ""}, nil }
	r := f.responses[f.idx]
	f.idx++
	return r, nil
}
func (f *fakeProv) Stream(ctx context.Context, req provider.CompletionRequest, onChunk func(provider.Chunk) error) error { return nil }
func (f *fakeProv) ListModels(ctx context.Context) ([]provider.Model, error) { return []provider.Model{{ID: f.model, Owner: "fake"}}, nil }

func TestRunSingleTurnNoTools(t *testing.T) {
	prov := &fakeProv{model: "m1", responses: []*provider.CompletionResponse{{Content: "hello"}}}
	exec := &Executor{Registry: tools.NewRegistry(), Auto: true}
	steps := []Step{}
	err := Run(context.Background(), prov, exec, "sys", 3, func(s Step) error { steps = append(steps, s); return nil })
	if err != nil { t.Fatalf("Run error: %v", err) }
	if len(steps) == 0 || steps[0].Content != "hello" { t.Fatalf("expected hello step, got: %+v", steps) }
}
