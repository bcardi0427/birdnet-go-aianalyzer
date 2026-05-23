package llm

import (
	"fmt"
	"strings"

	"github.com/tphakala/birdnet-go/internal/conf"
	"github.com/tphakala/birdnet-go/internal/logger"
)

// NewProvider creates a provider adapter based on active AI settings.
//
// Phase 3 introduces the abstraction and routing. Concrete provider adapters
// are implemented in Phase 4.
func NewProvider(settings conf.AISettings, apiKey string, log logger.Logger) (Provider, error) {
	providerID := normalizeProviderID(settings.Provider)

	if log == nil {
		log = logger.Global().Module("ai.llm")
	}

	switch providerID {
	case ProviderGemini:
		return newGeminiProvider(apiKey, log)
	case ProviderOpenAI:
		return newOpenAIProvider(apiKey, log)
	case ProviderOpenRouter:
		return newOpenRouterProvider(apiKey, log)
	case ProviderOpenAICompatible:
		return newOpenAICompatibleProvider(settings.BaseURL, apiKey, log)
	case ProviderOllama:
		return newOllamaProvider(settings.BaseURL, apiKey, log)
	case ProviderAnthropic:
		return newAnthropicProvider(apiKey, log)
	default:
		return nil, fmt.Errorf("unsupported AI provider %q", providerID)
	}
}

func normalizeProviderID(provider string) string {
	normalized := strings.ToLower(strings.TrimSpace(provider))
	if normalized == "" {
		return ProviderGemini
	}
	return normalized
}
