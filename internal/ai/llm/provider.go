package llm

import "context"

const (
	ProviderGemini           = "gemini"
	ProviderOpenAI           = "openai"
	ProviderOpenRouter       = "openrouter"
	ProviderOpenAICompatible = "openai-compatible"
	ProviderOllama           = "ollama"
	ProviderAnthropic        = "anthropic"
)

// Provider is the provider-neutral LLM contract used by AI report generation.
type Provider interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
	ListModels(ctx context.Context) ([]Model, error)
}

// GenerateRequest contains provider-agnostic prompt inputs.
type GenerateRequest struct {
	SystemPrompt string
	Prompt       string
	Model        string
}

// GenerateResponse contains provider-agnostic output text.
type GenerateResponse struct {
	Text string
}

// Model describes an available model from a provider.
type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}
