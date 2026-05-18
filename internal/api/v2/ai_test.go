package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tphakala/birdnet-go/internal/conf"
)

// TestGetAISettings_RedactsAPIKey verifies that GET /api/v2/ai/settings
// returns the API key redacted (replaced with **********) when a key is configured.
func TestGetAISettings_RedactsAPIKey(t *testing.T) {
	controller := newMinimalController()
	controller.Settings.AI.APIKey = "test-gemini-key-12345"
	controller.Settings.AI.Model = "gemini-2.5-flash"
	controller.Settings.AI.Enabled = true
	controller.Settings.AI.CacheHours = 4

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/ai/settings", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := controller.GetAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result conf.AISettings
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, redactedValue, result.APIKey, "API key must be redacted")
	assert.Equal(t, "gemini-2.5-flash", result.Model, "Model should be preserved")
	assert.True(t, result.Enabled, "Enabled should be preserved")
	assert.Equal(t, 4, result.CacheHours, "CacheHours should be preserved")
}

// TestGetAISettings_EmptyKeyNotRedacted verifies that GET /api/v2/ai/settings
// returns an empty string (not redacted) when the API key is empty.
func TestGetAISettings_EmptyKeyNotRedacted(t *testing.T) {
	controller := newMinimalController()
	controller.Settings.AI.APIKey = ""
	controller.Settings.AI.Model = "gemini-2.5-flash"
	controller.Settings.AI.Enabled = false

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v2/ai/settings", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := controller.GetAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result conf.AISettings
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Empty(t, result.APIKey, "Empty API key should remain empty, not redacted")
	assert.Equal(t, "gemini-2.5-flash", result.Model)
}

// TestUpdateAISettings_PreservesKeyWhenRedacted verifies that PATCH /api/v2/ai/settings
// preserves the current (real) API key when the incoming request contains the redacted placeholder.
func TestUpdateAISettings_PreservesKeyWhenRedacted(t *testing.T) {
	originalKey := "original-gemini-key-abc123"
	controller := newMinimalController()
	controller.Settings.AI.APIKey = originalKey
	controller.Settings.AI.Model = "gemini-2.5-flash"
	controller.Settings.AI.Enabled = true
	controller.Settings.AI.CacheHours = 4

	// Simulate a PATCH request with redacted API key
	update := conf.AISettings{
		APIKey:     redactedValue,
		Model:      "gemini-2.0-pro",
		Enabled:    true,
		CacheHours: 8,
	}

	body, err := json.Marshal(update)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v2/ai/settings", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err = controller.UpdateAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result conf.AISettings
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	// Verify response has redacted key
	assert.Equal(t, redactedValue, result.APIKey)
	assert.Equal(t, "gemini-2.0-pro", result.Model)
	assert.Equal(t, 8, result.CacheHours)

	// Verify the real key was preserved in the controller's settings
	assert.Equal(t, originalKey, controller.Settings.AI.APIKey,
		"Original key must be preserved when receiving redacted placeholder")
	assert.Equal(t, "gemini-2.0-pro", controller.Settings.AI.Model)
	assert.Equal(t, 8, controller.Settings.AI.CacheHours)
}

// TestUpdateAISettings_SavesNewKey verifies that PATCH /api/v2/ai/settings
// saves a new API key when a non-redacted value is provided.
func TestUpdateAISettings_SavesNewKey(t *testing.T) {
	controller := newMinimalController()
	controller.Settings.AI.APIKey = "old-key"
	controller.Settings.AI.Model = "gemini-2.5-flash"
	controller.Settings.AI.Enabled = true

	newKey := "new-gemini-key-xyz789"
	update := conf.AISettings{
		APIKey:     newKey,
		Model:      "gemini-2.0-pro",
		Enabled:    true,
		CacheHours: 6,
	}

	body, err := json.Marshal(update)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v2/ai/settings", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err = controller.UpdateAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result conf.AISettings
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	// Verify response has redacted key
	assert.Equal(t, redactedValue, result.APIKey)

	// Verify the new key was saved
	assert.Equal(t, newKey, controller.Settings.AI.APIKey,
		"New API key must be saved")
	assert.Equal(t, "gemini-2.0-pro", controller.Settings.AI.Model)
	assert.Equal(t, 6, controller.Settings.AI.CacheHours)
}

// TestUpdateAISettings_ClearsKey verifies that PATCH /api/v2/ai/settings
// can clear (empty) the API key when an empty string is provided.
func TestUpdateAISettings_ClearsKey(t *testing.T) {
	controller := newMinimalController()
	controller.Settings.AI.APIKey = "existing-key"
	controller.Settings.AI.Model = "gemini-2.5-flash"
	controller.Settings.AI.Enabled = true

	update := conf.AISettings{
		APIKey:     "",
		Model:      "gemini-2.0-pro",
		Enabled:    false,
		CacheHours: 4,
	}

	body, err := json.Marshal(update)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v2/ai/settings", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err = controller.UpdateAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result conf.AISettings
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	// Verify response has empty key
	assert.Empty(t, result.APIKey)

	// Verify the key was cleared in settings
	assert.Empty(t, controller.Settings.AI.APIKey)
	assert.False(t, controller.Settings.AI.Enabled)
}

// TestUpdateAISettings_FailsValidation_EnabledWithoutKey verifies that PATCH /api/v2/ai/settings
// returns 400 Bad Request when AI is enabled but API key is empty.
func TestUpdateAISettings_FailsValidation_EnabledWithoutKey(t *testing.T) {
	controller := newMinimalController()
	controller.Settings.AI.APIKey = ""
	controller.Settings.AI.Model = "gemini-2.5-flash"
	controller.Settings.AI.Enabled = false

	// Try to enable AI without providing a key
	update := conf.AISettings{
		APIKey:     "",
		Model:      "gemini-2.0-pro",
		Enabled:    true,
		CacheHours: 4,
	}

	body, err := json.Marshal(update)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v2/ai/settings", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err = controller.UpdateAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code,
		"Should return 400 when AI is enabled without API key")

	var errorResponse map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, "AI settings validation failed", errorResponse["message"])
	assert.NotEmpty(t, errorResponse["errors"],
		"Should include validation errors in response")

	errors, ok := errorResponse["errors"].([]any)
	assert.True(t, ok, "errors should be a list")
	assert.Greater(t, len(errors), 0, "Should have at least one error")
	assert.True(t, contains(errors, "Gemini API key is required when AI is enabled (set ai.apiKey or ai.apiKeyFile)"),
		"Error list should mention missing API key")
}

// TestUpdateAISettings_FailsValidation_EnabledWithoutModel verifies that PATCH /api/v2/ai/settings
// returns 400 Bad Request when AI is enabled but model is empty.
func TestUpdateAISettings_FailsValidation_EnabledWithoutModel(t *testing.T) {
	controller := newMinimalController()
	controller.Settings.AI.APIKey = "test-key"
	controller.Settings.AI.Model = ""
	controller.Settings.AI.Enabled = false

	// Try to enable AI without providing a model
	update := conf.AISettings{
		APIKey:     "test-key",
		Model:      "",
		Enabled:    true,
		CacheHours: 4,
	}

	body, err := json.Marshal(update)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v2/ai/settings", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err = controller.UpdateAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code,
		"Should return 400 when AI is enabled without model")

	var errorResponse map[string]any
	err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Equal(t, "AI settings validation failed", errorResponse["message"])
	errors, ok := errorResponse["errors"].([]any)
	assert.True(t, ok)
	assert.Greater(t, len(errors), 0)
	assert.True(t, contains(errors, "Gemini model is required when AI is enabled"),
		"Error list should mention missing model")
}

// TestUpdateAISettings_DisabledAllowsMissingKey verifies that PATCH /api/v2/ai/settings
// succeeds when AI is disabled, even with missing API key or model (no validation required).
func TestUpdateAISettings_DisabledAllowsMissingKey(t *testing.T) {
	controller := newMinimalController()
	controller.Settings.AI.APIKey = "old-key"
	controller.Settings.AI.Model = "gemini-2.5-flash"
	controller.Settings.AI.Enabled = true

	// Disable AI without key or model
	update := conf.AISettings{
		APIKey:     "",
		Model:      "",
		Enabled:    false,
		CacheHours: 4,
	}

	body, err := json.Marshal(update)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v2/ai/settings", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err = controller.UpdateAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code,
		"Should succeed when AI is disabled, even without key or model")

	var result conf.AISettings
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.False(t, result.Enabled)
	assert.Empty(t, result.APIKey)
	assert.Empty(t, result.Model)
}

// TestUpdateAISettings_InvalidCacheHours verifies that invalid CacheHours
// is normalized to 4 (with warning) and doesn't fail validation.
func TestUpdateAISettings_InvalidCacheHours(t *testing.T) {
	controller := newMinimalController()
	controller.Settings.AI.APIKey = "test-key"
	controller.Settings.AI.Model = "gemini-2.5-flash"
	controller.Settings.AI.Enabled = true

	// Try to set invalid cache hours (too low)
	update := conf.AISettings{
		APIKey:     "test-key",
		Model:      "gemini-2.0-pro",
		Enabled:    true,
		CacheHours: 0, // Invalid: must be at least 1
	}

	body, err := json.Marshal(update)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v2/ai/settings", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err = controller.UpdateAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result conf.AISettings
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	// Validation normalizes invalid CacheHours to 4
	assert.Equal(t, 4, result.CacheHours,
		"Invalid CacheHours should be normalized to 4")
}

// TestUpdateAISettings_InvalidRequestBody verifies that PATCH /api/v2/ai/settings
// returns 400 Bad Request for malformed JSON.
func TestUpdateAISettings_InvalidRequestBody(t *testing.T) {
	controller := newMinimalController()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v2/ai/settings",
		bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := controller.UpdateAISettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code,
		"Should return 400 for malformed JSON")
}

// Helper function to check if a list of errors contains a specific message
func contains(errors []any, msg string) bool {
	for _, e := range errors {
		if str, ok := e.(string); ok && str == msg {
			return true
		}
	}
	return false
}
