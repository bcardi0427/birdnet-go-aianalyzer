// conf/validate_ai.go

package conf

import (
	"fmt"
	"strings"

	"github.com/tphakala/birdnet-go/internal/errors"
	"github.com/tphakala/birdnet-go/internal/logger"
)

const (
	aiProviderGemini           = "gemini"
	aiProviderOpenAI           = "openai"
	aiProviderOpenRouter       = "openrouter"
	aiProviderOpenAICompatible = "openai-compatible"
	aiProviderOllama           = "ollama"
	aiProviderAnthropic        = "anthropic"

	ollamaDefaultBaseURL = "http://localhost:11434/v1"
)

var validAIProviders = map[string]struct{}{
	aiProviderGemini:           {},
	aiProviderOpenAI:           {},
	aiProviderOpenRouter:       {},
	aiProviderOpenAICompatible: {},
	aiProviderOllama:           {},
	aiProviderAnthropic:        {},
}

var aiProviderList = []string{
	aiProviderGemini,
	aiProviderOpenAI,
	aiProviderOpenRouter,
	aiProviderOpenAICompatible,
	aiProviderOllama,
	aiProviderAnthropic,
}

func normalizeAIProvider(provider string) string {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if normalized == "" {
		return aiProviderGemini
	}
	return normalized
}

func providerRequiresAPIKey(provider string) bool {
	return provider != aiProviderOllama
}

// ValidateAISettings performs AI settings validation without side effects.
// Returns validation result with normalized settings.
func ValidateAISettings(settings *AISettings) ValidationResult {
	if settings == nil {
		return ValidationResult{Valid: false, Errors: []string{"AI settings is nil"}}
	}
	result := ValidationResult{Valid: true, Warnings: []string{}}
	normalized := *settings
	normalized.Provider = normalizeAIProvider(settings.Provider)
	normalized.BaseURL = strings.TrimSpace(settings.BaseURL)

	if _, ok := validAIProviders[normalized.Provider]; !ok {
		result.Valid = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("AI provider must be one of: %s", strings.Join(aiProviderList, ", ")))
	}

	// ReportDays range: 1..31 (daily to monthly-ish window).
	if settings.ReportDays < 1 {
		result.Warnings = append(result.Warnings, "AI report days must be at least 1, defaulting to 1")
		normalized.ReportDays = 1
	} else if settings.ReportDays > 31 {
		result.Warnings = append(result.Warnings, "AI report days must be at most 31, capping to 31")
		normalized.ReportDays = 31
	}

	if settings.Enabled {
		// API key source is required by most providers when enabled.
		if providerRequiresAPIKey(normalized.Provider) &&
			strings.TrimSpace(settings.APIKey) == "" && strings.TrimSpace(settings.APIKeyFile) == "" {
			result.Valid = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("AI provider API key is required for %s when AI is enabled (set ai.apiKey or ai.apiKeyFile)", normalized.Provider))
		}

		// Model is required when enabled
		if strings.TrimSpace(settings.Model) == "" {
			result.Valid = false
			result.Errors = append(result.Errors, "AI model is required when AI is enabled")
		}

		if normalized.Provider == aiProviderOpenAICompatible && normalized.BaseURL == "" {
			result.Valid = false
			result.Errors = append(result.Errors,
				"AI base URL is required for openai-compatible when AI is enabled")
		}

		if normalized.Provider == aiProviderOllama && normalized.BaseURL == "" {
			normalized.BaseURL = ollamaDefaultBaseURL
		}

		// CacheHours should be at least 1, default to 4 if invalid
		if settings.CacheHours < 1 {
			result.Warnings = append(result.Warnings, "AI cache hours must be at least 1, defaulting to 4")
			normalized.CacheHours = 4
		}
	}

	result.Normalized = &normalized
	return result
}

// validateAISettings validates the AI-specific settings.
// This function uses ValidateAISettings internally and handles side effects
// (logging, mutation) to maintain backward compatibility.
func validateAISettings(settings *AISettings) error {
	result := ValidateAISettings(settings)

	normalized, err := extractNormalized[AISettings](result, "ValidateAISettings")
	if err != nil {
		return err
	}
	*settings = *normalized

	// Handle warnings (side effect: logging)
	for _, warning := range result.Warnings {
		GetLogger().Warn("AI validation warning", logger.String("message", warning))
	}

	// Return errors if validation failed
	if !result.Valid {
		return errors.Newf("ai settings errors: %v", strings.Join(result.Errors, "; ")).
			Category(errors.CategoryValidation).
			Context("validation_type", "ai-settings").
			Build()
	}

	return nil
}
