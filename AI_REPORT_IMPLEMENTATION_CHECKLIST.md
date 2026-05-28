# BirdNET-Go AI Report Implementation Checklist

This checklist is the working task list for implementing the AI daily analysis report described in [`ai_feature_blueprint.md`](ai_feature_blueprint.md).

## Core Implementation Rule

- [x] Keep the boundary clear: **Backend owns facts, tables, links, and images. Gemini owns narrative interpretation only.**
- [x] Do not ask Gemini to generate HTML tables, `<img>` tags, arbitrary HTML, or third-party image URLs.
- [x] Use BirdNET-Go local media endpoints for all report thumbnails.

---

## Phase 0: Pre-flight Review

- [x] Re-read `ai_feature_blueprint.md` before starting implementation.
- [x] Review existing AI settings implementation:
  - [x] `internal/conf/config.go`
  - [x] `internal/conf/validate_ai.go`
  - [x] `internal/api/v2/ai.go`
  - [x] `frontend/src/lib/desktop/features/settings/pages/AISettingsPage.svelte`
  - [x] `frontend/src/lib/utils/settingsApi.ts`
- [ ] Review existing media/species image endpoints:
  - [ ] `GET /api/v2/media/species-image?name={SCIENTIFIC_NAME}`
  - [ ] `GET /api/v2/media/species-image/info?name={SCIENTIFIC_NAME}`
  - [ ] `GET /api/v2/media/image/{SCIENTIFIC_NAME}`
- [x] Review existing weather APIs and datastore weather entities.
- [x] Review existing eBird integration and settings.

---

## Phase 1: Backend Route Foundation

- [x] Confirm AI routes are initialized in `internal/api/v2/ai.go`.
- [x] Add or complete `GET /api/v2/ai/report`.
- [x] Add or complete `GET /api/v2/ai/models` if not already complete.
- [ ] Ensure AI endpoints return safe, structured JSON responses.
- [ ] Ensure disabled AI settings return a clear, user-friendly error.
- [ ] Ensure missing Gemini API key returns a clear, user-friendly error.
- [x] Add request timeout handling for Gemini model calls.
- [ ] Add logging that does **not** include API keys, prompts with secrets, or sensitive user data.

---

## Phase 2: Report Data Aggregation

### Detection Aggregation

- [x] Query detections for the report window, initially last 24 hours.
- [x] Compute total detections.
- [x] Compute unique species count.
- [x] Compute high-confidence detection count and percentage.
- [x] Compute confidence buckets:
  - [x] 95–100%
  - [x] 90–94%
  - [x] 80–89%
  - [x] Below 80%
- [x] Compute hourly detection counts.
- [x] Identify peak activity hour/window.
- [x] Identify quietest activity window.
- [x] Compute top detected species.
- [x] For top species, compute:
  - [x] Common name
  - [x] Scientific name
  - [x] Detection count
  - [x] Average confidence
  - [x] Peak time window
  - [x] First seen / last seen timestamps if useful
- [x] Identify rare/notable detections using local database frequency.
- [ ] Identify unusual time-of-day detections only if supported by data/rules.

### Weather Aggregation

- [x] Pull weather data for the report period from existing weather storage/API.
- [x] Summarize available weather fields only:
  - [x] Temperature min/max/average
  - [x] Wind average/max if available
  - [ ] Precipitation/rain if available
  - [x] Humidity if available
  - [x] Weather condition labels/icons if available
- [x] If weather data is missing, mark it as unavailable rather than guessing.

### eBird Enrichment

- [x] Use existing eBird settings/API integration.
- [x] Add eBird context only when configured and available.
- [ ] Determine whether detected species are recently expected/reported in the configured region.
- [x] Generate approved eBird links from backend logic, not from Gemini.
- [ ] If eBird data is missing/unavailable, omit or label as unavailable.

---

## Phase 3: Backend-Generated Report Sections

- [x] Create a report data model for deterministic report assembly.
- [x] Create backend HTML escaping helpers for table content.
- [x] Generate deterministic HTML table for **Daily Totals**.
- [x] Generate deterministic HTML table for **Top Species**.
- [x] Generate deterministic HTML table for **Rare / Notable Detections**.
- [x] Generate deterministic HTML table for **Weather Summary**.
- [x] Generate deterministic HTML table for **Confidence Breakdown**.
- [x] Generate thumbnail cells using local BirdNET-Go media endpoints only:
  - [x] Use `/api/v2/media/species-image?name={SCIENTIFIC_NAME}`.
  - [x] URL-encode scientific names.
  - [x] Include useful `alt` text.
  - [x] Include `loading="lazy"`.
  - [x] Include fixed `width` and `height`.
- [ ] Add fallback behavior when no thumbnail is available.
- [x] Ensure backend-generated links use `rel="noopener noreferrer"` when opening a new tab.

---

## Phase 4: Gemini Narrative Generation

- [x] Build a structured context payload for Gemini using only backend-computed facts.
- [ ] Keep prompts concise and bounded to avoid excessive token use.
- [ ] System prompt must instruct Gemini:
  - [ ] Return markdown narrative only.
  - [ ] Do not generate HTML tables.
  - [ ] Do not generate `<img>` tags.
  - [ ] Do not provide image URLs.
  - [ ] Do not invent metrics.
  - [ ] Say “not available” or omit claims when data is missing.
  - [ ] Keep observations grounded in provided facts.
- [ ] Ask Gemini for narrative sections such as:
  - [ ] Bird activity overview.
  - [ ] Activity trend interpretation.
  - [ ] Weather correlation commentary.
  - [ ] Notable observations summary.
  - [ ] Short natural-language summary.
- [ ] Add timeout and error handling around Gemini calls.
- [ ] Add graceful fallback report when Gemini fails:
  - [ ] Return backend-generated tables.
  - [ ] Include a message that AI narrative is unavailable.

---

## Phase 5: Guardrails and Sanitization

### Backend Guardrails

- [x] Post-process Gemini narrative before assembling the final report.
- [x] Strip or reject any `<img>` tags in Gemini output.
- [x] Strip or reject image URLs from Gemini output.
- [x] Strip or reject unsafe HTML from Gemini output.
- [ ] Ensure final backend-generated HTML contains only approved tags/attributes.
- [ ] Ensure all user/data-derived strings in HTML are escaped.
- [ ] Ensure no API keys or secrets can appear in report context or output.

### Frontend Guardrails

- [x] Render markdown using `marked`.
- [x] Sanitize rendered output using `DOMPurify`.
- [ ] Configure DOMPurify allowlist for backend-generated report HTML:
  - [x] `table`, `thead`, `tbody`, `tr`, `th`, `td`, `caption`
  - [x] `a`
  - [x] `img`
- [ ] Allow only safe attributes:
  - [x] `href`, `title`, `rel`, `target`
  - [x] `src`, `alt`, `loading`, `width`, `height`
- [x] Consider restricting image `src` to local `/api/v2/media/` paths.

---

## Phase 6: Report Caching

- [x] Use configured `CacheHours` for report cache lifetime.
- [x] Cache the final assembled report, not just Gemini narrative.
- [x] Include timestamp metadata in cache.
- [ ] Add force-refresh support if desired later.
- [ ] Ensure cache is invalidated when relevant AI settings change:
  - [ ] Model changes.
  - [ ] System prompt changes.
  - [ ] Cache hours changes.
  - [ ] AI enabled/disabled changes.
- [ ] Ensure failed Gemini calls do not poison the cache unless intentionally caching fallback output.
  - [x] Current behavior: fallback reports from Gemini failure are returned but not cached.

---

## Phase 7: Frontend UI

- [ ] Create or complete `AIAnalysisPage.svelte`.
- [ ] Fetch report from `GET /api/v2/ai/report`.
- [ ] Show loading state while report is generated.
- [ ] Show cached/generated timestamp if provided by backend.
- [ ] Show useful error states:
  - [ ] AI disabled.
  - [ ] Missing API key.
  - [ ] Gemini unavailable.
  - [ ] No detections available.
- [ ] Render sanitized report content.
- [ ] Add a refresh button if backend supports refresh.
- [ ] Add page to navigation/sidebar.
- [ ] Ensure responsive layout for tables on small screens.
- [ ] Ensure report tables are readable in existing Tailwind/daisyUI theme.

---

## Phase 8: Settings UI

- [ ] Verify AI settings page supports:
  - [ ] Enabled toggle.
  - [ ] Gemini API key.
  - [ ] Model dropdown.
  - [ ] Cache hours.
  - [ ] System prompt.
- [ ] Verify API key redaction round-trip works.
- [ ] Verify model fetching requires an API key and handles errors cleanly.
- [ ] Verify settings validation rejects invalid cache hours/model values.

---

## Phase 9: Tests

> Note: test/build completion is currently blocked by Windows CGO linker path handling for `go test` in this environment (`invalid flag in go:cgo_ldflag` / `0xc0000135` seen during attempts).
> Build validation update: frontend production build and backend `go build -o birdnet-go.exe .` completed successfully after setting CGO/TFLite environment variables.

### Backend Tests

- [ ] Test detection aggregation with known fixture data.
- [ ] Test confidence bucket calculations.
- [ ] Test hourly peak/quiet window calculations.
- [ ] Test weather summary with complete data.
- [ ] Test weather summary with missing data.
- [ ] Test eBird enrichment enabled/disabled/unavailable.
- [ ] Test backend HTML table escaping.
- [ ] Test thumbnail URL generation uses local endpoints only.
- [ ] Test Gemini narrative guardrails strip image tags/URLs.
- [ ] Test cache hit/cache miss behavior.
- [ ] Test AI disabled/missing API key errors.

### Frontend Tests

- [x] Test report fetch success state.
- [ ] Test report loading state.
- [x] Test report error states.
- [x] Test sanitized markdown/HTML rendering.
- [ ] Test DOMPurify removes unsafe HTML.
- [ ] Test responsive table behavior if there are existing frontend test patterns for it.

---

## Phase 10: Manual Validation

- [ ] Start BirdNET-Go locally.
- [ ] Configure Gemini API key.
- [ ] Fetch available Gemini models.
- [ ] Generate a report with real or seeded detection data.
- [ ] Verify report includes backend-generated tables.
- [ ] Verify Gemini narrative does not contain invented metrics.
- [ ] Verify thumbnails load from local BirdNET-Go endpoints only.
- [ ] Verify no third-party image URLs appear in final report HTML.
- [ ] Verify report still works when weather is unavailable.
- [ ] Verify report still works when eBird is unavailable.
- [ ] Verify cache behavior by reloading within `CacheHours`.

---

## Phase 11: Documentation / Follow-up

- [ ] Update `ai_feature_blueprint.md` if implementation decisions change.
- [ ] Document any new AI report API response fields.
- [ ] Document report limitations:
  - [ ] AI narrative is interpretive.
  - [ ] Metrics are only included when supported by BirdNET-Go data.
  - [ ] Images use BirdNET-Go controlled endpoints.
- [ ] Add future enhancement notes:
  - [ ] 7-day / 30-day trend reports.
  - [ ] Rare bird/anomaly alert worker.
  - [ ] Report persona/tone setting.
  - [ ] Optional AI-generated first-party illustrations with clear labeling.

---

## Definition of Done for v1

- [ ] User can configure Gemini AI settings.
- [ ] User can generate a daily AI report from the web UI.
- [ ] Backend aggregates real BirdNET-Go facts for the report.
- [ ] Backend generates deterministic tables and local thumbnail URLs.
- [ ] Gemini generates narrative markdown only.
- [ ] Report rendering is sanitized.
- [ ] No AI-provided image URLs are accepted or displayed.
- [ ] Weather context is included when available.
- [ ] eBird context is included when available.
- [ ] Report gracefully degrades when Gemini/weather/eBird data is unavailable.
- [ ] Targeted backend and frontend tests pass.