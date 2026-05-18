// internal/api/v2/ai.go
package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tphakala/birdnet-go/internal/ai"
	"github.com/tphakala/birdnet-go/internal/conf"
	"github.com/tphakala/birdnet-go/internal/datastore/v2/repository"
	"github.com/tphakala/birdnet-go/internal/logger"
	"github.com/tphakala/birdnet-go/internal/secrets"
	"google.golang.org/genai"
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

	// Create auth-protected AI API group
	aiGroup := c.Group.Group("/ai", c.authMiddleware)

	// Routes for AI
	// GET /api/v2/ai/settings - Retrieves AI integration settings
	aiGroup.GET("/settings", c.GetAISettings)
	// PATCH /api/v2/ai/settings - Updates AI integration settings
	aiGroup.PATCH("/settings", c.UpdateAISettings)
	// GET /api/v2/ai/models - Retrieves available Gemini models
	aiGroup.GET("/models", c.GetAIModels)
	// GET /api/v2/ai/report - Generates or retrieves the daily AI report
	aiGroup.GET("/report", c.GetAIReport)

	c.logInfoIfEnabled("AI routes initialized successfully")
}

// GetAISettings handles GET /api/v2/ai/settings
func (c *Controller) GetAISettings(ctx echo.Context) error {
	c.logInfoIfEnabled("Getting AI settings",
		logger.String("path", ctx.Request().URL.Path),
		logger.String("ip", ctx.RealIP()),
	)

	c.settingsMutex.RLock()
	defer c.settingsMutex.RUnlock()

	settings := c.Settings.AI

	// Redact the API key
	settings.APIKey = redact(settings.APIKey)

	return ctx.JSON(http.StatusOK, settings)
}

// UpdateAISettings handles PATCH /api/v2/ai/settings
func (c *Controller) UpdateAISettings(ctx echo.Context) error {
	c.logInfoIfEnabled("Updating AI settings",
		logger.String("path", ctx.Request().URL.Path),
		logger.String("ip", ctx.RealIP()),
	)

	var update conf.AISettings
	if err := ctx.Bind(&update); err != nil {
		return c.HandleError(ctx, err, "Invalid request body", http.StatusBadRequest)
	}

	c.settingsMutex.Lock()
	defer c.settingsMutex.Unlock()

	current := c.getSettingsOrFallback()
	if current == nil {
		return c.HandleError(ctx, fmt.Errorf("settings not initialized"), "Failed to get settings", http.StatusInternalServerError)
	}

	keyAction := "saved"
	if update.APIKey == redactedValue {
		keyAction = "preserved"
		update.APIKey = current.AI.APIKey
	} else if update.APIKey == "" {
		keyAction = "cleared"
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

	updatedSettings := conf.CloneSettings(current)
	updatedSettings.AI = update
	c.Settings = updatedSettings
	conf.StoreSettings(updatedSettings)

	// Save settings to disk
	if err := conf.SaveSettings(); err != nil {
		return c.HandleError(ctx, err, "Failed to save settings", http.StatusInternalServerError)
	}

	GetLogger().Info("AI settings saved",
		logger.String("gemini_api_key", keyAction),
		logger.Bool("api_key_configured", update.APIKey != ""),
		logger.Bool("enabled", update.Enabled),
		logger.String("model", update.Model),
		logger.Int("cache_hours", update.CacheHours),
		logger.Bool("api_key_file_configured", strings.TrimSpace(update.APIKeyFile) != ""),
	)

	// Return the updated settings (redacted)
	update.APIKey = redact(update.APIKey)

	return ctx.JSON(http.StatusOK, update)
}

// GetAIModels handles GET /api/v2/ai/models
func (c *Controller) GetAIModels(ctx echo.Context) error {
	c.logInfoIfEnabled("Fetching Gemini models",
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

	if apiKey == "" {
		return c.HandleError(ctx, nil, "AI API Key is not configured", http.StatusBadRequest)
	}

	genaiCtx, cancel := context.WithTimeout(ctx.Request().Context(), 20*time.Second)
	defer cancel()
	client, err := genai.NewClient(genaiCtx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return c.HandleError(ctx, err, "Failed to create Gemini client", http.StatusInternalServerError)
	}

	var models []map[string]string
	modelPage, err := client.Models.List(genaiCtx, nil)
	if err != nil {
		return c.HandleError(ctx, err, "Failed to list Gemini models", http.StatusInternalServerError)
	}

	for {
		for _, model := range modelPage.Items {
			models = append(models, map[string]string{
				"id":          model.Name,
				"displayName": model.DisplayName,
				"description": model.Description,
			})
		}

		nextPage, err := modelPage.Next(genaiCtx)
		if err == genai.ErrPageDone {
			break
		}
		if err != nil {
			return c.HandleError(ctx, err, "Failed to fetch models from Gemini", http.StatusInternalServerError)
		}

		modelPage = nextPage
	}

	return ctx.JSON(http.StatusOK, models)
}

// GetAIReport handles GET /api/v2/ai/report
func (c *Controller) GetAIReport(ctx echo.Context) error {
	c.logInfoIfEnabled("Generating AI report",
		logger.String("path", ctx.Request().URL.Path),
		logger.String("ip", ctx.RealIP()),
	)

	if c.aiService == nil {
		return c.HandleError(ctx, nil, "AI service not initialized", http.StatusInternalServerError)
	}

	report, err := c.aiService.GetDailyReport(ctx.Request().Context())
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			return c.HandleError(ctx, err, "AI report generation timed out", http.StatusGatewayTimeout)
		case strings.Contains(errMsg, "disabled"):
			return c.HandleError(ctx, err, "AI analysis is disabled", http.StatusBadRequest)
		case strings.Contains(errMsg, "api key"):
			return c.HandleError(ctx, err, "Gemini API key is not configured", http.StatusBadRequest)
		default:
			return c.HandleError(ctx, err, "Failed to generate AI report", http.StatusInternalServerError)
		}
	}

	return ctx.JSON(http.StatusOK, report)
}
