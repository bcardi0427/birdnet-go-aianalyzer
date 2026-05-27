// internal/api/v2/ai.go
package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tphakala/birdnet-go/internal/ai"
	"github.com/tphakala/birdnet-go/internal/ai/llm"
	"github.com/tphakala/birdnet-go/internal/api/auth"
	"github.com/tphakala/birdnet-go/internal/conf"
	"github.com/tphakala/birdnet-go/internal/datastore/v2/repository"
	interrors "github.com/tphakala/birdnet-go/internal/errors"
	"github.com/tphakala/birdnet-go/internal/logger"
	"github.com/tphakala/birdnet-go/internal/secrets"
)

// initAIRoutes registers all AI-related API endpoints
func (c *Controller) initAIRoutes() {
	c.logInfoIfEnabled("Initializing AI routes")

	// Lazily initialize the AI report service from the v2 detection repository.
	// This mirrors the insightsRepo pattern: V2Manager may not be set during
	// NewWithOptions (e.g., in unit tests), so we defer initialization here.
	if c.aiService == nil && c.V2Manager != nil {
		db := c.V2Manager.DB()
		useV2Prefix := c.V2Manager.TablePrefix() != ""
		isMySQL := c.V2Manager.IsMySQL()
		detRepo := repository.NewDetectionRepository(db, nil, useV2Prefix, isMySQL)
		weatherRepo := repository.NewWeatherRepository(db, nil, useV2Prefix, isMySQL)
		labelRepo := repository.NewLabelRepository(db, nil, useV2Prefix, isMySQL)
		c.aiService = ai.NewReportService(c.Settings, detRepo, weatherRepo, labelRepo, c.EBirdClient)
	}

	// Report reads are public so visitors can view the cached AI summary.
	// Cache bypass remains protected inside GetAIReport because it can spend AI tokens.
	aiGroup := c.Group.Group("/ai")

	// GET /api/v2/ai/report - Generates or retrieves the daily AI report
	aiGroup.GET("/report", c.GetAIReport)

	protectedAIGroup := c.Group.Group("/ai", c.authMiddleware)
	// GET /api/v2/ai/settings - Retrieves AI integration settings
	protectedAIGroup.GET("/settings", c.GetAISettings)
	// PATCH /api/v2/ai/settings - Updates AI integration settings
	protectedAIGroup.PATCH("/settings", c.UpdateAISettings)
	// GET /api/v2/ai/models - Retrieves available models for active provider
	protectedAIGroup.GET("/models", c.GetAIModels)

	c.logInfoIfEnabled("AI routes initialized successfully")
}

// GetAISettings retrieves the current AI settings with API keys redacted.
func (c *Controller) GetAISettings(ctx echo.Context) error {
	c.logInfoIfEnabled("Getting AI settings",
		logger.String("path", ctx.Request().URL.Path),
		logger.String("ip", ctx.RealIP()),
	)

	c.settingsMutex.RLock()
	defer c.settingsMutex.RUnlock()

	settings := c.Settings.AI

	// Redact all API keys
	settings.APIKey = redact(settings.APIKey)
	settings.Gemini.APIKey = redact(settings.Gemini.APIKey)
	settings.OpenAI.APIKey = redact(settings.OpenAI.APIKey)
	settings.OpenRouter.APIKey = redact(settings.OpenRouter.APIKey)
	settings.OpenAICompatible.APIKey = redact(settings.OpenAICompatible.APIKey)
	settings.Ollama.APIKey = redact(settings.Ollama.APIKey)
	settings.Anthropic.APIKey = redact(settings.Anthropic.APIKey)

	return ctx.JSON(http.StatusOK, settings)
}

// UpdateAISettings updates the AI settings, supporting legacy and new formats,
// validating the settings and storing them.
func (c *Controller) UpdateAISettings(ctx echo.Context) error {
	c.logInfoIfEnabled("Updating AI settings",
		logger.String("path", ctx.Request().URL.Path),
		logger.String("ip", ctx.RealIP()),
	)

	var update conf.AISettings
	if err := ctx.Bind(&update); err != nil {
		return c.HandleError(ctx, err, "Invalid request body", http.StatusBadRequest)
	}

	GetLogger().Info("AI settings PATCH payload received",
		logger.String("provider_raw", update.Provider),
		logger.String("model_raw", update.Model),
		logger.Bool("enabled_raw", update.Enabled),
		logger.Bool("api_key_is_redacted", update.APIKey == redactedValue),
		logger.Bool("api_key_present", strings.TrimSpace(update.APIKey) != ""),
	)

	c.settingsMutex.Lock()
	defer c.settingsMutex.Unlock()

	current := c.getSettingsOrFallback()
	if current == nil {
		return c.HandleError(ctx, interrors.Newf("settings not initialized").
			Component("api").
			Category(interrors.CategorySystem).
			Build(), "Failed to get settings", http.StatusInternalServerError)
	}

	currentProvider := strings.TrimSpace(strings.ToLower(current.AI.Provider))
	incomingProvider := strings.TrimSpace(strings.ToLower(update.Provider))
	// Preserve provider on partial/legacy payloads that omit it.
	if incomingProvider == "" {
		incomingProvider = currentProvider
		update.Provider = current.AI.Provider
	}
	providerChanged := incomingProvider != currentProvider

	// Check if this is a legacy client by seeing if all provider structures are zero/empty
	isLegacy := (update.Gemini.APIKey == "" && update.Gemini.APIKeyFile == "" && update.Gemini.Model == "" &&
		update.OpenAI.APIKey == "" && update.OpenAI.APIKeyFile == "" && update.OpenAI.Model == "" &&
		update.OpenRouter.APIKey == "" && update.OpenRouter.APIKeyFile == "" && update.OpenRouter.Model == "" &&
		update.OpenAICompatible.APIKey == "" && update.OpenAICompatible.APIKeyFile == "" && update.OpenAICompatible.Model == "" &&
		update.Ollama.APIKey == "" && update.Ollama.APIKeyFile == "" && update.Ollama.Model == "" &&
		update.Anthropic.APIKey == "" && update.Anthropic.APIKeyFile == "" && update.Anthropic.Model == "")

	keyAction := "saved"

	if isLegacy {
		keyAction = restoreLegacyAISettings(&update, &current.AI, providerChanged, incomingProvider)
	} else {
		restoreNewAISettings(&update, &current.AI)
	}

	if strings.TrimSpace(update.BaseURL) == "" && strings.TrimSpace(current.AI.BaseURL) != "" {
		// Keep existing base URL unless caller explicitly changes it.
		// This avoids accidental resets when frontends submit partial payloads.
		update.BaseURL = current.AI.BaseURL
	}

	// Validate after redacted placeholders are restored so a configured API key
	// can survive a GET → PATCH round trip when AI is enabled.
	validationResult := conf.ValidateAISettings(&update)
	if !validationResult.Valid {
		c.logInfoIfEnabled("AI settings validation failed",
			logger.String("errors", strings.Join(validationResult.Errors, "; ")),
		)
		return ctx.JSON(http.StatusBadRequest, map[string]any{
			"message": "AI settings validation failed",
			"errors":  validationResult.Errors,
		})
	}
	if normalized, ok := validationResult.Normalized.(*conf.AISettings); ok && normalized != nil {
		update = *normalized
	}

	GetLogger().Info("AI settings normalized before save",
		logger.String("provider_normalized", update.Provider),
		logger.String("model_normalized", update.Model),
		logger.Bool("enabled_normalized", update.Enabled),
		logger.Bool("api_key_configured_normalized", strings.TrimSpace(update.APIKey) != ""),
	)

	updatedSettings := conf.CloneSettings(current)
	updatedSettings.AI = update
	GetLogger().Info("AI settings to be persisted",
		logger.String("provider_persist", updatedSettings.AI.Provider),
		logger.String("model_persist", updatedSettings.AI.Model),
		logger.Bool("enabled_persist", updatedSettings.AI.Enabled),
	)
	c.Settings = updatedSettings
	conf.StoreSettings(updatedSettings)

	// Save settings to disk
	if err := conf.SaveSettings(); err != nil {
		return c.HandleError(ctx, err, "Failed to save settings", http.StatusInternalServerError)
	}

	GetLogger().Info("AI settings saved",
		logger.String("ai_api_key", keyAction),
		logger.String("provider", update.Provider),
		logger.Bool("api_key_configured", update.APIKey != ""),
		logger.Bool("enabled", update.Enabled),
		logger.String("model", update.Model),
		logger.Int("report_days", update.ReportDays),
		logger.Int("cache_hours", update.CacheHours),
		logger.Bool("api_key_file_configured", strings.TrimSpace(update.APIKeyFile) != ""),
	)

	// Return the updated settings (redacted)
	update.APIKey = redact(update.APIKey)
	update.Gemini.APIKey = redact(update.Gemini.APIKey)
	update.OpenAI.APIKey = redact(update.OpenAI.APIKey)
	update.OpenRouter.APIKey = redact(update.OpenRouter.APIKey)
	update.OpenAICompatible.APIKey = redact(update.OpenAICompatible.APIKey)
	update.Ollama.APIKey = redact(update.Ollama.APIKey)
	update.Anthropic.APIKey = redact(update.Anthropic.APIKey)

	return ctx.JSON(http.StatusOK, update)
}

// GetAIModels fetches the list of available models from the configured AI provider.
func (c *Controller) GetAIModels(ctx echo.Context) error {
	c.logInfoIfEnabled("Fetching AI provider models",
		logger.String("path", ctx.Request().URL.Path),
		logger.String("ip", ctx.RealIP()),
	)

	c.settingsMutex.RLock()
	aiSettings := c.Settings.AI
	c.settingsMutex.RUnlock()

	apiKey, source, resolveErr := secrets.ResolveWithSource(aiSettings.APIKeyFile, aiSettings.APIKey)
	if resolveErr != nil {
		return c.HandleError(ctx, resolveErr, "Failed to resolve AI API key", http.StatusBadRequest)
	}
	if source == secrets.SecretSourceEnvOrText && !secrets.IsEnvReference(aiSettings.APIKey) && aiSettings.APIKey != "" {
		GetLogger().Warn("plaintext secret in use; migrate to env var or secret file",
			logger.String("field", "ai.apiKey"),
			logger.String("source", "plaintext"),
		)
	}

	providerID := strings.TrimSpace(strings.ToLower(aiSettings.Provider))
	if providerID == "" {
		providerID = llm.ProviderGemini
	}

	if apiKey == "" && providerID != llm.ProviderOllama {
		return c.HandleError(ctx, nil, "AI API Key is not configured", http.StatusBadRequest)
	}

	provider, err := llm.NewProvider(aiSettings, apiKey, GetLogger().Module("api.ai.models"))
	if err != nil {
		return c.HandleError(ctx, err, "Failed to create AI provider client", http.StatusInternalServerError)
	}

	listCtx, cancel := context.WithTimeout(ctx.Request().Context(), 20*time.Second)
	defer cancel()
	providerModels, err := provider.ListModels(listCtx)
	if err != nil {
		fallbackModel := strings.TrimSpace(aiSettings.Model)
		if fallbackModel == "" {
			switch providerID {
			case llm.ProviderOpenAI:
				fallbackModel = llm.DefaultOpenAIModel
			case llm.ProviderOpenRouter:
				fallbackModel = llm.DefaultOpenRouterModel
			case llm.ProviderOllama:
				fallbackModel = llm.DefaultOllamaModel
			case llm.ProviderAnthropic:
				fallbackModel = llm.DefaultAnthropicModel
			default:
				fallbackModel = llm.DefaultGeminiModel
			}
		}
		return ctx.JSON(http.StatusOK, []llm.Model{{ID: fallbackModel, DisplayName: fallbackModel, Description: "Configured/default model"}})
	}
	return ctx.JSON(http.StatusOK, providerModels)
}

// GetAIReport generates or retrieves the daily bird detection report.
func (c *Controller) GetAIReport(ctx echo.Context) error {
	c.logInfoIfEnabled("Generating AI report",
		logger.String("path", ctx.Request().URL.Path),
		logger.String("ip", ctx.RealIP()),
	)

	if c.aiService == nil {
		return c.HandleError(ctx, nil, "AI service not initialized", http.StatusInternalServerError)
	}

	bypassCache := false
	if raw := strings.TrimSpace(ctx.QueryParam("bypass_cache")); raw != "" {
		if parsed, parseErr := strconv.ParseBool(raw); parseErr == nil {
			bypassCache = parsed
		}
	}

	if bypassCache && !c.isExplicitlyAuthenticated(ctx) {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Login required to bypass the AI report cache",
		})
	}

	report, err := c.aiService.GetDailyReport(ctx.Request().Context(), bypassCache)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return c.HandleError(ctx, err, "AI report generation timed out", http.StatusGatewayTimeout)
		case strings.Contains(errMsg, "disabled"):
			return c.HandleError(ctx, err, "AI analysis is disabled", http.StatusBadRequest)
		case strings.Contains(errMsg, "api key"):
			return c.HandleError(ctx, err, "AI provider API key is not configured", http.StatusBadRequest)
		default:
			return c.HandleError(ctx, err, "Failed to generate AI report", http.StatusInternalServerError)
		}
	}

	return ctx.JSON(http.StatusOK, report)
}

// isExplicitlyAuthenticated checks if the request is authenticated with basic auth,
// token, oauth, browser session, or API key.
func (c *Controller) isExplicitlyAuthenticated(ctx echo.Context) bool {
	if c.authService != nil && !c.authService.IsAuthRequired(ctx) {
		return true
	}

	method, ok := ctx.Get(auth.CtxKeyAuthMethod).(auth.AuthMethod)
	if !ok {
		return false
	}

	switch method {
	case auth.AuthMethodBasicAuth, auth.AuthMethodToken, auth.AuthMethodOAuth2, auth.AuthMethodBrowserSession, auth.AuthMethodAPIKey:
		return true
	default:
		return false
	}
}

// restoreLegacyAISettings handles restoring redacted root API keys and synchronizing
// settings for legacy clients. It returns the action taken on the key.
func restoreLegacyAISettings(update, current *conf.AISettings, providerChanged bool, incomingProvider string) string {
	keyAction := "saved"
	if update.APIKey == redactedValue {
		if providerChanged {
			keyAction = "cleared_on_provider_change"
			update.APIKey = ""
		} else {
			keyAction = "preserved"
			update.APIKey = current.APIKey
		}
	} else if update.APIKey == "" {
		keyAction = "cleared"
	}

	update.MigrateAndSync(false)

	if incomingProvider != "gemini" { update.Gemini = current.Gemini }
	if incomingProvider != "openai" { update.OpenAI = current.OpenAI }
	if incomingProvider != "openrouter" { update.OpenRouter = current.OpenRouter }
	if incomingProvider != "openai-compatible" { update.OpenAICompatible = current.OpenAICompatible }
	if incomingProvider != "ollama" { update.Ollama = current.Ollama }
	if incomingProvider != "anthropic" { update.Anthropic = current.Anthropic }

	return keyAction
}

// restoreNewAISettings restores redacted API keys for each provider and handles
// merging empty provider structures for newer clients.
func restoreNewAISettings(update, current *conf.AISettings) {
	if update.Gemini.APIKey == redactedValue {
		update.Gemini.APIKey = current.Gemini.APIKey
	}
	if update.OpenAI.APIKey == redactedValue {
		update.OpenAI.APIKey = current.OpenAI.APIKey
	}
	if update.OpenRouter.APIKey == redactedValue {
		update.OpenRouter.APIKey = current.OpenRouter.APIKey
	}
	if update.OpenAICompatible.APIKey == redactedValue {
		update.OpenAICompatible.APIKey = current.OpenAICompatible.APIKey
	}
	if update.Ollama.APIKey == redactedValue {
		update.Ollama.APIKey = current.Ollama.APIKey
	}
	if update.Anthropic.APIKey == redactedValue {
		update.Anthropic.APIKey = current.Anthropic.APIKey
	}

	isEmpty := func(p conf.AIProviderSettings) bool {
		return p.APIKey == "" && p.APIKeyFile == "" && p.BaseURL == "" && p.Model == ""
	}
	if isEmpty(update.Gemini) { update.Gemini = current.Gemini }
	if isEmpty(update.OpenAI) { update.OpenAI = current.OpenAI }
	if isEmpty(update.OpenRouter) { update.OpenRouter = current.OpenRouter }
	if isEmpty(update.OpenAICompatible) { update.OpenAICompatible = current.OpenAICompatible }
	if isEmpty(update.Ollama) { update.Ollama = current.Ollama }
	if isEmpty(update.Anthropic) { update.Anthropic = current.Anthropic }

	update.MigrateAndSync(false)
}
