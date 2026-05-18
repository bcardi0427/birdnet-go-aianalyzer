# Secrets Hardening Implementation Report

Date: 2026-05-18  
Branch: `feature/ai-daily-report`

## Summary
Implemented secrets hardening for AI, weather, and eBird settings by adding `apiKeyFile` support, extending secret resolution to return source metadata, updating runtime consumers to resolve from file-or-value safely, and fixing AI settings update normalization behavior.

## Commits
- `78e13820` `harden secrets: support apiKeyFile sources and normalize AI settings updates`
- `34c062c6` `docs: convert secrets hardening plan into checklist`

## Files Changed
- `internal/conf/config.go`
- `internal/conf/defaults.go`
- `internal/conf/validate_ai.go`
- `internal/conf/validate_realtime.go`
- `internal/secrets/secrets.go`
- `internal/api/v2/ai.go`
- `internal/api/v2/ai_test.go`
- `internal/api/v2/integrations.go`
- `internal/weather/provider_openweather.go`
- `internal/weather/provider_wunderground.go`
- `internal/conf/config_yaml_tags_test.go`
- `internal/conf/validate_hardening_test.go`
- `plan/secrets-hardening-plan.md`

## Detailed Changes

### 1) Config schema: `apiKeyFile` fields
Added `apiKeyFile` to settings structs so users can source secrets from files instead of plaintext config:
- `Realtime.EBird.APIKeyFile`
- `Realtime.Weather.OpenWeather.APIKeyFile`
- `Realtime.Weather.Wunderground.APIKeyFile`
- `AI.APIKeyFile`

Also updated default values to include empty `apikeyfile` keys where applicable.

### 2) Validation hardening
- AI validation now requires at least one source when enabled:
  - `ai.apiKey` or `ai.apiKeyFile`
- Realtime validation updated similarly:
  - eBird requires `apiKey` or `apiKeyFile` when enabled
  - OpenWeather provider requires `apiKey` or `apiKeyFile`
- Wunderground validation accepts `apiKey` or `apiKeyFile`.

### 3) Secrets resolver extension
Extended secrets utility with:
- `ResolveWithSource(filePath, value) (secret, source, err)`
- `SecretSource` indicator values (none/file/env-or-text)
- `IsEnvReference(value)` helper for `${...}` detection

This allows callers to both resolve secret values and understand where they came from.

### 4) Runtime integration updates
Updated runtime consumers to use `ResolveWithSource`:
- AI model listing route (`/api/v2/ai/models`)
- OpenWeather provider
- Wunderground provider
- eBird client initialization path
- weather integration auth test path in API v2 integrations

Added warning logs for plaintext secret usage to support migration to env vars/secret files.

### 5) AI settings update regression fix
In `UpdateAISettings` (`internal/api/v2/ai.go`):
- Apply normalized output from `ValidateAISettings` before saving.
- Ensures invalid cache hours are normalized (e.g. `0 -> 4`) and persisted correctly.

### 6) Tests updated
- Added/updated YAML round-trip coverage for new `apiKeyFile` fields.
- Updated validation tests for weather/eBird requirements.
- Added AI route tests and updated expectation text for new key-source validation wording.

## Verification Performed
User-ran and confirmed passing:
- `go test github.com/tphakala/birdnet-go/internal/api/v2 -run TestUpdateAISettings -count=1 -v`
- `go test github.com/tphakala/birdnet-go/internal/api/v2 -run "TestGetAISettings|TestUpdateAISettings" -count=1 -v`

Both passed after applying the normalization + assertion fixes.

## Notes
- Repository has additional unrelated in-progress changes in working tree; this work was committed in scoped commits.
- Windows line-ending warnings (`LF -> CRLF`) were observed during staging; no functional impact to logic.
