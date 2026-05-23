package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tphakala/birdnet-go/internal/logger"
)

type anthropicProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	log     logger.Logger
}

func newAnthropicProvider(apiKey string, log logger.Logger) (Provider, error) {
	return &anthropicProvider{apiKey: strings.TrimSpace(apiKey), baseURL: DefaultAnthropicBaseURL, client: http.DefaultClient, log: log}, nil
}

func (p *anthropicProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	payload := map[string]any{
		"model":      effectiveModel(req.Model, DefaultAnthropicModel),
		"max_tokens": 1024,
		"system":     req.SystemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": req.Prompt},
		},
	}

	var response struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := p.doJSON(ctx, http.MethodPost, "/messages", payload, &response); err != nil {
		return GenerateResponse{}, err
	}

	parts := make([]string, 0, len(response.Content))
	for _, item := range response.Content {
		if item.Type == "text" && strings.TrimSpace(item.Text) != "" {
			parts = append(parts, strings.TrimSpace(item.Text))
		}
	}
	return GenerateResponse{Text: strings.Join(parts, "\n\n")}, nil
}

func (p *anthropicProvider) ListModels(_ context.Context) ([]Model, error) {
	return []Model{{ID: DefaultAnthropicModel, DisplayName: DefaultAnthropicModel, Description: "Anthropic default model"}}, nil
}

func (p *anthropicProvider) doJSON(ctx context.Context, method, path string, payload, target any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Anthropic-Version", "2023-06-01")
	if p.apiKey != "" {
		req.Header.Set("x-api-key", p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("anthropic request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
