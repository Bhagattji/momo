package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	sharedClient     *http.Client
	sharedClientOnce sync.Once
)

func getSharedClient() *http.Client {
	sharedClientOnce.Do(func() {
		transport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		}
		sharedClient = &http.Client{
			Transport: transport,
		}
	})
	return sharedClient
}

type OpenAICompat struct {
	name    string
	model   string
	apiKey  string
	baseURL string
}

func (o *OpenAICompat) Name() string    { return o.name }
func (o *OpenAICompat) Model() string   { return o.model }
func (o *OpenAICompat) SetModel(m string) { o.model = m }

func doRequestWithRetry(ctx context.Context, req *http.Request, maxAttempts int) (*http.Response, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
	}
	client := getSharedClient()
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		resp, err := client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if attempt == maxAttempts {
				return nil, err
			}
			sleep := time.Duration(1<<uint(attempt-1))*time.Second + time.Duration(rand.Intn(500))*time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(sleep):
			}
			continue
		}
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			ra := resp.Header.Get("Retry-After")
			_ = resp.Body.Close()
			if attempt == maxAttempts {
				return nil, fmt.Errorf("provider error: %s", resp.Status)
			}
			var wait time.Duration
			if secs, err := strconv.Atoi(ra); err == nil {
				wait = time.Duration(secs) * time.Second
			} else if t, err := http.ParseTime(ra); err == nil {
				wait = time.Until(t)
			}
			if wait <= 0 {
				wait = time.Duration(1<<uint(attempt-1)) * time.Second
			}
			wait += time.Duration(rand.Intn(500)) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("exhausted retries")
}

type oaMessage struct {
	Role       string       `json:"role"`
	Content    string       `json:"content,omitempty"`
	ToolCalls  []oaToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
}

type oaToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type oaTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	} `json:"function"`
}

func (o *OpenAICompat) buildMessages(req CompletionRequest) []oaMessage {
	out := make([]oaMessage, 0, len(req.Messages)+1)
	if req.System != "" {
		out = append(out, oaMessage{Role: "system", Content: req.System})
	}
	for _, m := range req.Messages {
		om := oaMessage{Role: m.Role, Content: m.Content, ToolCallID: m.ToolCallID}
		for _, tc := range m.ToolCalls {
			var c oaToolCall
			c.ID = tc.ID
			c.Type = "function"
			c.Function.Name = tc.Name
			c.Function.Arguments = tc.Arguments
			om.ToolCalls = append(om.ToolCalls, c)
		}
		out = append(out, om)
	}
	return out
}

func (o *OpenAICompat) buildTools(req CompletionRequest) []oaTool {
	if len(req.Tools) == 0 {
		return nil
	}
	out := make([]oaTool, 0, len(req.Tools))
	for _, t := range req.Tools {
		var ot oaTool
		ot.Type = "function"
		ot.Function.Name = t.Name
		ot.Function.Description = t.Description
		ot.Function.Parameters = t.Parameters
		out = append(out, ot)
	}
	return out
}

func (o *OpenAICompat) Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	endpoint := o.baseURL
	if endpoint == "" {
		endpoint = "https://api.openai.com"
	}
	url := strings.TrimRight(endpoint, "/") + "/chat/completions"

	payload := map[string]any{
		"model":    o.model,
		"messages": o.buildMessages(req),
		"stream":   false,
	}
	if tools := o.buildTools(req); tools != nil {
		payload["tools"] = tools
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	hreq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("User-Agent", "momo/"+version())
	if o.apiKey != "" {
		hreq.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := doRequestWithRetry(ctx, hreq, 3)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("provider error: %s: %s", resp.Status, stripError(body))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content   string       `json:"content"`
				ToolCalls []oaToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return &CompletionResponse{Content: string(body)}, nil
	}

	out := &CompletionResponse{}
	out.Usage.InputTokens = parsed.Usage.PromptTokens
	out.Usage.OutputTokens = parsed.Usage.CompletionTokens
	out.Usage.TotalTokens = parsed.Usage.TotalTokens
	if len(parsed.Choices) > 0 {
		out.Content = parsed.Choices[0].Message.Content
		for _, tc := range parsed.Choices[0].Message.ToolCalls {
			out.ToolCalls = append(out.ToolCalls, ToolCall{
				ID: tc.ID, Name: tc.Function.Name, Arguments: tc.Function.Arguments,
			})
		}
	}
	return out, nil
}

func (o *OpenAICompat) Stream(ctx context.Context, req CompletionRequest, onChunk func(Chunk) error) error {
	endpoint := o.baseURL
	if endpoint == "" {
		endpoint = "https://api.openai.com"
	}
	url := strings.TrimRight(endpoint, "/") + "/chat/completions"

	payload := map[string]any{
		"model":    o.model,
		"messages": o.buildMessages(req),
		"stream":   true,
	}
	if tools := o.buildTools(req); tools != nil {
		payload["tools"] = tools
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	hreq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("User-Agent", "momo/v1")
	if o.apiKey != "" {
		hreq.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := doRequestWithRetry(ctx, hreq, 3)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("provider error: %s: %s", resp.Status, stripError(body))
	}

	type accTool struct {
		ID, Name, Args string
	}
	pending := map[int]*accTool{}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var chunkObj struct {
			Choices []struct {
				Delta struct {
					Content   string       `json:"content"`
					ToolCalls []oaToolCall `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunkObj); err != nil {
			continue
		}
		if len(chunkObj.Choices) == 0 {
			continue
		}
		delta := chunkObj.Choices[0].Delta
		if delta.Content != "" {
			if err := onChunk(Chunk{Content: delta.Content}); err != nil {
				return err
			}
		}
		for i, tc := range delta.ToolCalls {
			a := pending[i]
			if a == nil {
				a = &accTool{}
				pending[i] = a
			}
			if tc.ID != "" {
				a.ID = tc.ID
			}
			if tc.Function.Name != "" {
				a.Name = tc.Function.Name
			}
			a.Args += tc.Function.Arguments
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if len(pending) > 0 {
		var calls []ToolCall
		for i := 0; i < len(pending); i++ {
			if a := pending[i]; a != nil && a.Name != "" {
				calls = append(calls, ToolCall{ID: a.ID, Name: a.Name, Arguments: a.Args})
			}
		}
		if len(calls) > 0 {
			if err := onChunk(Chunk{ToolCalls: calls, Done: true}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *OpenAICompat) ListModels(ctx context.Context) ([]Model, error) {
	if ctx == nil {
		return []Model{{ID: o.model, Owner: o.name}}, fmt.Errorf("nil context")
	}
	endpoint := o.baseURL
	if endpoint == "" {
		endpoint = "https://api.openai.com"
	}
	url := strings.TrimRight(endpoint, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "momo/v1")
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}
	client := getSharedClient()
	resp, err := client.Do(req)
	if err != nil {
		return []Model{{ID: o.model, Owner: o.name}}, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return []Model{{ID: o.model, Owner: o.name}}, nil
	}
	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err := json.Unmarshal(body, &out); err != nil || len(out.Data) == 0 {
		return []Model{{ID: o.model, Owner: o.name}}, nil
	}
	models := make([]Model, 0, len(out.Data))
	for _, m := range out.Data {
		models = append(models, Model{ID: m.ID, Owner: o.name})
	}
	return models, nil
}

func stripError(b []byte) string {
	if len(b) > 1024 {
		return string(b[:1024])
	}
	return string(b)
}

func version() string {
	return "1.0.0"
}