package llm

import "strings"

const (
	DefaultGeminiModel       = "gemini-2.5-flash"
	DefaultOpenAIModel       = "gpt-4o-mini"
	DefaultOpenRouterModel   = "openai/gpt-4o-mini"
	DefaultOllamaModel       = "llama3.2"
	DefaultAnthropicModel    = "claude-3-5-haiku-latest"
	DefaultOpenAIBaseURL     = "https://api.openai.com/v1"
	DefaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"
	DefaultOllamaBaseURL     = "http://localhost:11434/v1"
	DefaultAnthropicBaseURL  = "https://api.anthropic.com/v1"
)

func effectiveModel(model, fallback string) string {
	if strings.TrimSpace(model) == "" {
		return fallback
	}
	return strings.TrimSpace(model)
}

func trimTrailingSlash(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}
