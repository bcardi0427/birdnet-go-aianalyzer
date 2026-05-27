package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAISettings_Phase2Normalization(t *testing.T) {
	t.Parallel()

	settings := &AISettings{
		Enabled:    false,
		Provider:   "  OpenAI  ",
		BaseURL:    "  https://api.example.com/v1  ",
		ReportDays: 1,
		CacheHours: 4,
	}

	result := ValidateAISettings(settings)
	assert.True(t, result.Valid, "expected provider normalization to remain valid")

	normalized, ok := result.Normalized.(*AISettings)
	require.True(t, ok)
	require.NotNil(t, normalized)
	assert.Equal(t, "openai", normalized.Provider)
	assert.Equal(t, "https://api.example.com/v1", normalized.BaseURL)
}

func TestValidateAISettings_DefaultProviderGeminiWhenEmpty(t *testing.T) {
	t.Parallel()

	settings := &AISettings{Enabled: false, Provider: ""}
	result := ValidateAISettings(settings)
	assert.True(t, result.Valid)

	normalized, ok := result.Normalized.(*AISettings)
	require.True(t, ok)
	require.NotNil(t, normalized)
	assert.Equal(t, "gemini", normalized.Provider)
}

func TestValidateAISettings_RejectsUnknownProvider(t *testing.T) {
	t.Parallel()

	settings := &AISettings{Enabled: false, Provider: "unknown-provider"}
	result := ValidateAISettings(settings)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors,
		"AI provider must be one of: gemini, openai, openrouter, openai-compatible, ollama, anthropic")
}

func TestValidateAISettings_RequiresAPIKeyForRemoteProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
	}{
		{name: "gemini", provider: "gemini"},
		{name: "openai", provider: "openai"},
		{name: "openrouter", provider: "openrouter"},
		{name: "openai-compatible", provider: "openai-compatible"},
		{name: "anthropic", provider: "anthropic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			settings := &AISettings{
				Enabled:    true,
				Provider:   tt.provider,
				APIKey:     "",
				APIKeyFile: "",
				Model:      "some-model",
				BaseURL:    "http://localhost:4000/v1",
				ReportDays: 1,
				CacheHours: 4,
			}

			result := ValidateAISettings(settings)
			assert.False(t, result.Valid)
			assert.Contains(t, result.Errors,
				"AI provider API key is required for "+tt.provider+" when AI is enabled (set ai.apiKey or ai.apiKeyFile)")
		})
	}
}

func TestValidateAISettings_OllamaDoesNotRequireAPIKey(t *testing.T) {
	t.Parallel()

	settings := &AISettings{
		Enabled:    true,
		Provider:   "ollama",
		APIKey:     "",
		APIKeyFile: "",
		Model:      "llama3.2",
		ReportDays: 1,
		CacheHours: 4,
	}

	result := ValidateAISettings(settings)
	assert.True(t, result.Valid, "ollama should allow keyless configuration")

	normalized, ok := result.Normalized.(*AISettings)
	require.True(t, ok)
	require.NotNil(t, normalized)
	assert.Equal(t, "http://localhost:11434/v1", normalized.BaseURL)
}

func TestValidateAISettings_OpenAICompatibleRequiresBaseURLWhenEnabled(t *testing.T) {
	t.Parallel()

	settings := &AISettings{
		Enabled:    true,
		Provider:   "openai-compatible",
		APIKey:     "test-key",
		Model:      "my-model",
		BaseURL:    "",
		ReportDays: 1,
		CacheHours: 4,
	}

	result := ValidateAISettings(settings)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors, "AI base URL is required for openai-compatible when AI is enabled")
}

func TestValidateAISettings_ModelRequiredWhenEnabled(t *testing.T) {
	t.Parallel()

	settings := &AISettings{
		Enabled:    true,
		Provider:   "openai",
		APIKey:     "test-key",
		Model:      "   ",
		ReportDays: 1,
		CacheHours: 4,
	}

	result := ValidateAISettings(settings)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors, "AI model is required when AI is enabled")
}
