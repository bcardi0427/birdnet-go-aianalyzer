# AI Analysis UI Polish Plan

Date: 2026-05-17

Scope: Polish the BirdNET-Go AI Analysis UI in the frontend with frontend-only UI changes and backend URL governance, without altering the underlying backend report generation logic.

Objective: Deliver a visually improved, robust, and mobile-friendly AI Analysis UI that preserves backend data semantics and security constraints.

Related security plan (must be tracked in parallel):
- `plan/secrets-hardening-plan.md`
- Reason: AI feature rollout introduces/uses API-key-driven integrations; secrets handling hardening is a release-critical parallel stream.

Critical constraints:
- Backend remains the source of truth for all data; Gemini only supplies narrative content.
- Do not allow Gemini to generate outbound URLs or HTML.
- All outbound eBird links must be augmented with safe UTMs on the backend; Gemini must not modify URLs.
- UI changes must be frontend-only; no backend logic changes beyond URL shaping/modest UI tweaks.

## Phase 1: Phase Kickoff & Requirements Trace
- [x] Confirm plan with product on permitted changes (UI polish only).
- [x] List touched files: backend: internal/ai/service.go; frontend: frontend/src/lib/desktop/features/ai/AIAnalysisPage.svelte, frontend/src/lib/desktop/features/ai/AIAnalysisPage.test.ts, frontend/src/lib/desktop/features/settings/pages/AISettingsPage.svelte.
- [x] Identify existing rendering points for the Rare / Notable Detections section and the Top Detections table.
- [x] Confirm existing URL generation patterns for eBird links and ensure URL encoding is preserved.

## Phase 2: Backend changes (URL governance and safe data)
Goals:
- [x] Append safe outbound UTMs to all backend-generated eBird URLs in the AI report (backend-controlled only).
- [x] Introduce a small internal helper to merge UTMs safely without duplicating existing parameters.
- [x] Add a lightweight external-link indicator at the frontend when rendering eBird links (↗) while preserving target="_blank" and rel="noopener noreferrer".
- [x] Ensure eBird URL construction keeps proper encoding and avoids duplicate query params.
- [x] Ensure skeleton loading in Refresh uses backend-determined data; UTMs do not affect skeletons.
Notes:
- Implement the UTMs in the backend at the point where EBirdURL is generated (buildSpeciesRows). Ensure we only append UTMs if missing from the URL and preserve existing query parameters.
- Do not rely on Gemini for outbound URL creation.

Files touched (high level):
- internal/ai/service.go: renderNotableSpecies, buildSpeciesRows, ebirdLinkHTML, imageHTML, plus a URL-UTM utility.
- internal/api/v2/ai.go: none (already wired for endpoint; outbound URLs originate in AI service).

## Phase 3: Frontend polish (UI-only changes)
Goals:
- Rare / Notable Detections visual polish: add a tasteful amber/star accent or badge for notable detections with low counts. Use DaisyUI/Tailwind-safe classes.
- External eBird link indicator: show a small ↗ icon next to the link; maintain safe link behavior.
- Skeleton loading during Refresh/loading: render skeleton blocks that resemble the final report: title block, a few summary rows, and a subset of table rows to imply structure.
- Responsive tables: ensure horizontal scroll wrappers for narrow viewports; avoid wrapping of long species names; use nowrap with ellipsis where needed; ensure readability on mobile.
- Tests: extend AIAnalysisPage.test.ts to cover the new badge rendering, external link indicator, and skeleton during loading.

Frontend changes (targeted):
- AIAnalysisPage.svelte: modify renderNotableSpecies to inject badge for low-count items, and ensure the eBird URLs render with an external indicator.
- AIAnalysisPage.svelte: wrap all tables with a responsive container (overflow-x-auto) for horizontal scrolling on small devices.
- AIAnalysisPage.test.ts: add tests for badge presence and external indicator in the rendered HTML, plus skeleton rendering state when loading.
- AISettingsPage.svelte: minor UI/UX support to reflect the new behavior without changing data flow.

## Phase 4: Testing & Validation
- [ ] Frontend unit tests for badge rendering and external link indicators.
- [ ] Frontend tests for skeleton loading state on Refresh.
- [ ] End-to-end sanity check by loading the AI Analysis page and verifying the URL structure and table rendering.
- [ ] Optional: manual QA in browser for small-width screens to confirm responsive behavior.

## Phase 5: Documentation & Review
- [ ] Update AI_REPORT_IMPLEMENTATION_CHECKLIST.md with completed items and attach a short notes section about UI polish.
- [ ] Document any UI design decisions in a short design note for future reference.

## Parallel Track: Secrets Hardening (Release-Critical)
- Coordinate implementation timing with `plan/secrets-hardening-plan.md`.
- Ensure AI settings and report endpoints do not regress secret redaction/safe persistence behavior.
- Treat this as a same-milestone dependency, not a post-release follow-up.

## Acceptance Criteria
- Frontend AI Analysis page uses UI-only polish as described without backend logic changes.
- Outbound eBird URLs include UTMs and a small external indicator; Gemini does not modify URLs.
- Skeleton loading state approximates final report while data loads.
- Tables remain readable on mobile and allow horizontal scrolling for wide content.

## Review & Diffs
- I will provide diffs for backend URL changes and frontend UI patches in subsequent commits.

## Plan File Path
- plan/ai_analysis_ui_plan.md
