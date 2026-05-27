package llm

import "strings"

const (
	// DefaultGeminiModel is the default model name for Gemini provider.
	DefaultGeminiModel       = "gemini-2.5-flash"
	// DefaultOpenAIModel is the default model name for OpenAI provider.
	DefaultOpenAIModel       = "gpt-4o-mini"
	// DefaultOpenRouterModel is the default model name for OpenRouter provider.
	DefaultOpenRouterModel   = "openai/gpt-4o-mini"
	// DefaultOllamaModel is the default model name for Ollama provider.
	DefaultOllamaModel       = "llama3.2"
	// DefaultAnthropicModel is the default model name for Anthropic provider.
	DefaultAnthropicModel    = "claude-3-5-haiku-latest"
	// DefaultOpenAIBaseURL is the default API base URL for OpenAI.
	DefaultOpenAIBaseURL     = "https://api.openai.com/v1"
	// DefaultOpenRouterBaseURL is the default API base URL for OpenRouter.
	DefaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"
	// DefaultOllamaBaseURL is the default API base URL for Ollama.
	DefaultOllamaBaseURL     = "http://localhost:11434/v1"
	// DefaultAnthropicBaseURL is the default API base URL for Anthropic.
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
