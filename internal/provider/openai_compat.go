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
	"time"
)

// OpenAICompat is a minimal compatible provider implementation that talks to
// OpenAI-compatible endpoints (OpenAI, OpenRouter, Groq, etc.).
type OpenAICompat struct {
	name    string
	model   string
	apiKey  string
	baseURL string
}

func (o *OpenAICompat) Name() string  { return o.name }
func (o *OpenAICompat) Model() string { return o.model }
func (o *OpenAICompat) SetModel(m string) { o.model = m }

// doRequestWithRetry performs HTTP requests with simple exponential backoff,
// honoring Retry-After header when present. Caller must close the response body.
func doRequestWithRetry(ctx context.Context, client *http.Client, req *http.Request, maxAttempts int) (*http.Response, error) {
	rand.Seed(time.Now().UnixNano())
	if maxAttempts < 1 { maxAttempts = 1 }
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Clone the request body if needed - for simplicity assume body is re-creatable by caller
		resp, err := client.Do(req)
		if err != nil {
			// network error - retry unless context done or last attempt
			if ctx.Err() != nil { return nil, ctx.Err() }
			if attempt == maxAttempts { return nil, err }
			// backoff with jitter
			sleep := time.Duration(1<<uint(attempt-1)) * time.Second
			sleep += time.Duration(rand.Intn(500)) * time.Millisecond
			select {
			case <-ctx.Done(): return nil, ctx.Err()
			case <-time.After(sleep):
			}
			continue
		}

		// If status suggests retry (rate limit or server error), handle Retry-After then retry
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			ra := resp.Header.Get("Retry-After")
			_ = resp.Body.Close()
			if attempt == maxAttempts { return nil, fmt.Errorf("provider error: %s", resp.Status) }
			// parse Retry-After header
			var wait time.Duration
			if ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil {
					wait = time.Duration(secs) * time.Second
				} else if t, err := http.ParseTime(ra); err == nil {
					wait = time.Until(t)
				}
			}
			if wait <= 0 {
				wait = time.Duration(1<<uint(attempt-1)) * time.Second
			}
			wait += time.Duration(rand.Intn(500)) * time.Millisecond
			select {
			case <-ctx.Done(): return nil, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		return resp, nil
	}
	return nil, fmt.Errorf("exhausted retries")
}

func (o *OpenAICompat) Chat(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	endpoint := o.baseURL
	if endpoint == "" {
		endpoint = "https://api.openai.com"
	}
	url := endpoint + "/v1/chat/completions"

	// Build payload
	types := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		types = append(types, map[string]string{"role": m.Role, "content": m.Content})
	}
	payload := map[string]interface{}{
		"model":    o.model,
		"messages": types,
		"stream":   req.Stream,
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
	if o.apiKey != "" {
		hreq.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := doRequestWithRetry(ctx, client, hreq, 5)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("provider error: %s: %s", resp.Status, string(body))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Text string `json:"text"`
		} `json:"choices"`
		Usage map[string]interface{} `json:"usage"`
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 65536))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		// If parsing fails, return raw body as content
		return &CompletionResponse{Content: string(body)}, nil
	}

	content := ""
	if len(parsed.Choices) > 0 {
		if parsed.Choices[0].Message.Content != "" {
			content = parsed.Choices[0].Message.Content
		} else {
			content = parsed.Choices[0].Text
		}
	}

	// Note: ToolCalls and Usage parsing omitted for now.
	return &CompletionResponse{Content: content}, nil
}

func (o *OpenAICompat) Stream(ctx context.Context, req CompletionRequest, onChunk func(Chunk) error) error {
	client := &http.Client{Timeout: 0}
	endpoint := o.baseURL
	if endpoint == "" {
		endpoint = "https://api.openai.com"
	}
	url := endpoint + "/v1/chat/completions"

	// Build payload with stream=true
	types := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		types = append(types, map[string]string{"role": m.Role, "content": m.Content})
	}
	payload := map[string]interface{}{
		"model":    o.model,
		"messages": types,
		"stream":   true,
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
	if o.apiKey != "" {
		hreq.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := doRequestWithRetry(ctx, client, hreq, 5)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("provider error: %s: %s", resp.Status, string(body))
	}

	// Read SSE-like streaming lines. Each data: line may be a JSON chunk.
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line == "\n" {
			continue
		}
		// Expect lines like: data: {"id":..., "choices": [...]}
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				// final
				return nil
			}
			var chunkObj map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunkObj); err != nil {
				// ignore non-JSON keep-alive
				continue
			}
			// Try to extract content from common fields
			content := ""
			if choices, ok := chunkObj["choices"].([]interface{}); ok && len(choices) > 0 {
				if ch, ok := choices[0].(map[string]interface{}); ok {
					// delta.content (OpenAI chunk) or message.content
					if delta, ok := ch["delta"].(map[string]interface{}); ok {
						if c, ok := delta["content"].(string); ok {
							content = c
						}
					}
					if content == "" {
						if msg, ok := ch["message"].(map[string]interface{}); ok {
							if c, ok := msg["content"].(string); ok { content = c }
						}
					}
					if content == "" {
						if text, ok := ch["text"].(string); ok { content = text }
					}
				}
			}
			if content != "" {
				if err := onChunk(Chunk{Content: content}); err != nil {
					return err
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (o *OpenAICompat) ListModels(ctx context.Context) ([]Model, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	endpoint := o.baseURL
	if endpoint == "" {
		endpoint = "https://api.openai.com"
	}
	url := endpoint + "/v1/models"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		// fallback to single default model
		return []Model{{ID: o.model, Owner: o.name}}, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return []Model{{ID: o.model, Owner: o.name}}, nil
	}
	var out struct {
		Data []struct{ ID string `json:"id"` } `json:"data"`
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 65536))
	if err := json.Unmarshal(body, &out); err != nil || len(out.Data) == 0 {
		return []Model{{ID: o.model, Owner: o.name}}, nil
	}
	models := make([]Model, 0, len(out.Data))
	for _, m := range out.Data {
		models = append(models, Model{ID: m.ID, Owner: o.name})
	}
	return models, nil
}
