## Plan: Finish AI Settings redaction flow

TL;DR: The AI backend and frontend integration is already implemented, but the remaining work is to add validation and dedicated tests around the AI settings save path so the feature is complete and safe.

### Goals
- Ensure `UpdateAISettings` validates AI settings before persisting.
- Preserve the real Gemini API key when the frontend sends the redacted placeholder.
- Verify GET and PATCH behavior with tests.
- Confirm the frontend uses the same sentinel and route wiring.

### Steps
1. Update backend logic in `internal/api/v2/ai.go`
   - Call `conf.ValidateAISettings(&update)` after binding the request and before cloning or saving settings.
   - Keep the existing redaction-preservation logic for `update.APIKey == redactedValue`.
   - Return a clear `400 Bad Request` with validation errors when AI settings are invalid.
2. Add backend API tests for AI settings endpoints.
   - Cover `GET /api/v2/ai/settings` returns redacted `apiKey` when set.
   - Cover `PATCH /api/v2/ai/settings` preserves the stored API key when the incoming value is `**********`.
   - Cover `PATCH /api/v2/ai/settings` saves a new API key when provided.
   - Cover validation failure when required AI fields are missing or invalid.
3. Add frontend test coverage or component validation.
   - Verify `AISettingsPage.svelte` loads the redacted placeholder and still allows saving without re-entering the API key.
   - Confirm the page calls `settingsAPI.ai.getModels()` only when the key is configured.
4. Sanity-check route and API client wiring.
   - Confirm `/ui/settings/ai` is registered in `internal/api/server.go` as a protected SPA route.
   - Confirm `frontend/src/lib/utils/settingsApi.ts` uses `/api/v2/ai/settings` for GET and PATCH.
5. Run targeted verification.
   - Run related Go tests for `internal/api/v2` and AI validation.
   - Run frontend settings tests if available.

### Relevant files
- `internal/api/v2/ai.go` — AI settings endpoint logic
- `internal/api/v2/settings.go` — redaction helpers and validation sentinel
- `internal/api/v2/settings_sanitize_test.go` — existing sanitization tests that can guide AI-specific tests
- `frontend/src/lib/desktop/features/settings/pages/AISettingsPage.svelte` — AI settings UI
- `frontend/src/lib/utils/settingsApi.ts` — frontend API client
- `internal/api/server.go` — SPA route registration

### Verification
1. `GET /api/v2/ai/settings` returns `apiKey: "**********"` when an API key exists.
2. `PATCH /api/v2/ai/settings` with `apiKey: "**********"` retains the stored real key.
3. `PATCH /api/v2/ai/settings` with a new key replaces the old key.
4. Validation errors are returned if AI is enabled without a key or model.
5. `AISettingsPage.svelte` does not leak the real API key in the UI and still functions after save.

### Decisions
- Reuse the existing redaction sentinel `**********` and backend helpers in `internal/api/v2/settings.go`.
- Keep AI settings endpoint implementation separate from the generic settings handler to minimize risk.
