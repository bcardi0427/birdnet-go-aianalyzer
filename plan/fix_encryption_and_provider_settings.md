# Implementation Plan - Configuration Encryption & LLM Provider Settings Fixes

This plan outlines the root causes and proposed changes to fix the following issues:
1. **Configuration & Storage Hardening:** Configuration secrets (e.g. SessionSecret) are written in plaintext to `config.yaml` during configuration migrations or session secret generation at startup.
2. **LLM Provider Reversion:** Non-Gemini LLM providers (like OpenAI) revert back to Gemini (with an empty API key) when saving other settings pages in the UI.

---

## Issue 1: Configuration & Storage Hardening (Plaintext Secrets)

### Root Cause
At startup, `conf.Load()` loads the configuration from `config.yaml` and decrypts its secrets in-memory using `decryptConfigSecrets()`. 

However, if a configuration migration triggers (e.g. `MigrateLocationConfigured`) or if a session secret is generated on first boot (`ensureSessionSecret`), the application writes the updated settings back to the configuration file. It does this by calling `SaveYAMLConfig(configFile, settings)` directly in `internal/conf/migrations.go`. Since the passed `settings` object is the in-memory version containing **decrypted** secrets, `SaveYAMLConfig` serializes and writes the secrets to `config.yaml` in plaintext.

### Proposed Changes

#### `internal/conf/migrations.go`
We will update the two functions that write configurations to disk to clone and encrypt the settings before saving.

1. **`persistMigration`**:
   - Clone the decrypted in-memory `settings` using `CloneSettings(settings)`.
   - Encrypt the clone's secrets in-place using `encryptConfigSecrets(settingsCopy)`.
   - Pass the encrypted clone to `SaveYAMLConfig`.
   
   ```go
   // persistMigration saves the config file after a successful migration.
   func persistMigration(settings *Settings, label string) {
   	configFile := viper.ConfigFileUsed()
   	if configFile == "" {
   		return
   	}
   	settingsCopy := CloneSettings(settings)
   	if err := encryptConfigSecrets(settingsCopy); err != nil {
   		GetLogger().Warn("Failed to encrypt configuration secrets for "+label, logger.Error(err))
   		return
   	}
   	if err := SaveYAMLConfig(configFile, settingsCopy); err != nil {
   		GetLogger().Warn("Failed to save migrated "+label+" config", logger.Error(err))
   	} else {
   		GetLogger().Info("Saved migrated "+label+" configuration", logger.String("path", configFile))
   	}
   }
   ```

2. **`ensureSessionSecret`**:
   - Similar to `persistMigration`, clone `settings` and encrypt the clone before passing it to `SaveYAMLConfig`.
   - The in-memory `settings` object remains updated with the decrypted version.
   
   ```go
   	// ... (after sessionSecret is generated and assigned to settings.Security.SessionSecret)
   
   	// Save the updated config back to file to persist the generated secret
   	// This ensures the secret remains the same across restarts
   	configFile := viper.ConfigFileUsed()
   	if configFile == "" {
   		return nil
   	}
   
   	settingsCopy := CloneSettings(settings)
   	if err := encryptConfigSecrets(settingsCopy); err != nil {
   		GetLogger().Warn("Failed to encrypt configuration secrets for SessionSecret save", logger.Error(err))
   		return nil
   	}
   
   	if err := SaveYAMLConfig(configFile, settingsCopy); err != nil {
   		// Log the error but don't fail - the generated secret will work for this session
   		GetLogger().Warn("Failed to save generated SessionSecret to config file", logger.Error(err))
   		return nil
   	}
   ```

---

## Issue 2: LLM Provider Reversion

### Root Cause
In `internal/conf/config.go`, the `RealtimeSettings` struct contains a duplicate `AI` field:
```go
type RealtimeSettings struct {
	...
	AI               AISettings               `yaml:"ai" json:"ai"`                             // AI features settings
}
```
This duplicate `Realtime.AI` field is completely unused by the Go backend (all features read from the root `Settings.AI` field). However, because of this duplicate field:
1. `GET /api/v2/settings` serializes both root `AI` and duplicate `Realtime.AI` fields. The root `AI.APIKey` is redacted correctly, but `Realtime.AI` is sent with empty/uninitialized fields.
2. The Svelte settings store loads this JSON response. The Svelte store does not define `ai` within `RealtimeSettings`, but the raw JS object retains it.
3. When saving any other settings page (e.g. general, audio, etc.), Svelte PUTs the entire settings object back to `/api/v2/settings`. This payload contains `"realtime": { "ai": { "provider": "gemini", "apiKey": "" } }`.
4. The Go backend unmarshals this payload into `updatedSettings`. In `updateAllowedSettingsWithTracking`, the backend recursively copies the empty `updatedSettings.Realtime.AI` fields into the active configuration structure.
5. In `viper` (which is case-insensitive and flattens keys), writing or unmarshaling `realtime.ai` conflicts with or overrides the root `ai` field, reverting the active provider to Gemini and clearing the API key.

### Proposed Changes

#### `internal/conf/config.go`
Delete the duplicate `AI` field from `RealtimeSettings`:
```diff
 type RealtimeSettings struct {
 	...
 	SpeciesTracking  SpeciesTrackingSettings  `yaml:"speciestracking" json:"speciesTracking"`   // New species tracking settings
 	ExtendedCapture  ExtendedCaptureSettings  `yaml:"extendedcapture" json:"extendedCapture"`   // Extended capture for long calling species
-	AI               AISettings               `yaml:"ai" json:"ai"`                             // AI features settings
 }
```

No code references `Realtime.AI` in either the backend Go codebase or the frontend Svelte application, so removing this duplicate field is safe and will resolve the reversion bug.

---

## Verification Plan

### Automated Tests
1. Run target unit tests in `internal/conf` to verify compilation and cloning correctness:
   `go test -v -run TestClone ./internal/conf/...`
2. Run settings API tests to ensure settings updates work without regressions:
   `go test -v ./internal/api/v2/...`
3. We can write a specific integration test checking that:
   - Config file migrations and `ensureSessionSecret` output files have all secrets (including `SessionSecret`) properly encrypted on disk (i.e. prefixed with `enc:v1:`).

### Manual Verification
1. Start the server, let migrations/session secret generation run, and verify `config.yaml` to ensure no plaintext secrets are saved.
2. Configure a non-Gemini LLM provider (e.g. OpenAI) with a mock API key via the UI AI Settings page, save it, and then modify another page (like the node name on the General Settings page) and save. Verify that the AI provider and API key do not revert to Gemini/empty.
