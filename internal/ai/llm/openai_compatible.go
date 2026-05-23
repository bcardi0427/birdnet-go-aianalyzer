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

type openAICompatibleProvider struct {
	providerID   string
	apiKey       string
	baseURL      string
	defaultModel string
	client       *http.Client
	log          logger.Logger
}

func newOpenAIProvider(apiKey string, log logger.Logger) (Provider, error) {
	return newOpenAICompatible(ProviderOpenAI, DefaultOpenAIBaseURL, apiKey, DefaultOpenAIModel, log), nil
}

func newOpenRouterProvider(apiKey string, log logger.Logger) (Provider, error) {
	return newOpenAICompatible(ProviderOpenRouter, DefaultOpenRouterBaseURL, apiKey, DefaultOpenRouterModel, log), nil
}

func newOpenAICompatibleProvider(baseURL, apiKey string, log logger.Logger) (Provider, error) {
	return newOpenAICompatible(ProviderOpenAICompatible, baseURL, apiKey, "", log), nil
}

func newOllamaProvider(baseURL, apiKey string, log logger.Logger) (Provider, error) {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultOllamaBaseURL
	}
	return newOpenAICompatible(ProviderOllama, baseURL, apiKey, DefaultOllamaModel, log), nil
}

func newOpenAICompatible(providerID, baseURL, apiKey, defaultModel string, log logger.Logger) *openAICompatibleProvider {
	return &openAICompatibleProvider{
		providerID:   providerID,
		apiKey:       strings.TrimSpace(apiKey),
		baseURL:      trimTrailingSlash(baseURL),
		defaultModel: defaultModel,
		client:       http.DefaultClient,
		log:          log,
	}
}

type openAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []openAIChatMessage `json:"messages"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIChatMessage `json:"message"`
	} `json:"choices"`
}

func (p *openAICompatibleProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	payload := openAIChatRequest{
		Model: effectiveModel(req.Model, p.defaultModel),
		Messages: []openAIChatMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.Prompt},
		},
	}

	var response openAIChatResponse
	if err := p.doJSON(ctx, http.MethodPost, "/chat/completions", payload, &response); err != nil {
		return GenerateResponse{}, err
	}
	if len(response.Choices) == 0 {
		return GenerateResponse{}, fmt.Errorf("%s returned no choices", p.providerID)
	}
	return GenerateResponse{Text: strings.TrimSpace(response.Choices[0].Message.Content)}, nil
}

func (p *openAICompatibleProvider) ListModels(ctx context.Context) ([]Model, error) {
	var response struct {
		Data []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"data"`
	}
	if err := p.doJSON(ctx, http.MethodGet, "/models", nil, &response); err != nil {
		fallback := effectiveModel("", p.defaultModel)
		if fallback == "" {
			return nil, err
		}
		return []Model{{ID: fallback, DisplayName: fallback, Description: p.providerID + " configured model"}}, nil
	}

	models := make([]Model, 0, len(response.Data))
	for _, item := range response.Data {
		displayName := item.Name
		if displayName == "" {
			displayName = item.ID
		}
		models = append(models, Model{ID: item.ID, DisplayName: displayName, Description: item.Description})
	}
	return models, nil
}

func (p *openAICompatibleProvider) doJSON(ctx context.Context, method, path string, payload, target any) error {
	if p.baseURL == "" {
		return fmt.Errorf("%s base URL is not configured", p.providerID)
	}

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("%s request failed with status %d: %s", p.providerID, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	if target == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}
