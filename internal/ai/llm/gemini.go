package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/tphakala/birdnet-go/internal/logger"
	"google.golang.org/genai"
)

type geminiProvider struct {
	apiKey string
	log    logger.Logger
}

func newGeminiProvider(apiKey string, log logger.Logger) (Provider, error) {
	return &geminiProvider{apiKey: strings.TrimSpace(apiKey), log: log}, nil
}

func (p *geminiProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: p.apiKey})
	if err != nil {
		return GenerateResponse{}, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	prompt := strings.TrimSpace(req.Prompt)
	if strings.TrimSpace(req.SystemPrompt) != "" {
		prompt = strings.TrimSpace(req.SystemPrompt) + "\n\n" + prompt
	}

	result, err := client.Models.GenerateContent(ctx, effectiveModel(req.Model, DefaultGeminiModel), genai.Text(prompt), nil)
	if err != nil {
		return GenerateResponse{}, err
	}

	return GenerateResponse{Text: strings.TrimSpace(result.Text())}, nil
}

func (p *geminiProvider) ListModels(ctx context.Context) ([]Model, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: p.apiKey})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	modelPage, err := client.Models.List(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list Gemini models: %w", err)
	}

	models := []Model{}
	for {
		for _, model := range modelPage.Items {
			models = append(models, Model{ID: model.Name, DisplayName: model.DisplayName, Description: model.Description})
		}

		nextPage, err := modelPage.Next(ctx)
		if err == genai.ErrPageDone {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Gemini models: %w", err)
		}
		modelPage = nextPage
	}

	return models, nil
}
