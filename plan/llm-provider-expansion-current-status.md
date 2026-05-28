# LLM Provider Expansion — Current Status and Execution Plan

## Current Position

We are **not** in Phase 1. Based on the implemented code and checklist state, we are effectively **post-Phase 8** and actively stabilizing behavior before/while completing **Phase 9 (testing)** and documentation.

Current checklist state in `plan/llm-provider-expansion-checklist.md` shows:

- Phase 1 ✅
- Phase 2 ✅
- Phase 3 ✅
- Phase 4 ✅
- Phase 5 ✅
- Phase 6 ✅
- Phase 7 ✅
- Phase 8 ✅
- Phase 9 ⏳ (open)
- Documentation ⏳ (open)

## What We Are Doing Right Now

We are fixing a **post-Phase-8 integration bug** in AI settings persistence and provider-specific behavior:

- Reported behavior indicated saves could revert to Gemini-like behavior when selecting another provider (notably OpenAI).
- Recent patch work focused on ensuring frontend save payloads are normalized and provider-consistent.

### Recent Active Fixes

In `frontend/src/lib/desktop/features/settings/pages/AISettingsPage.svelte`:

1. Normalize and always persist `payload.provider` on save.
2. Ensure provider-specific base URL defaults are preserved when appropriate (`openai`, `openrouter`, `ollama`).
3. Avoid stale Gemini model values when switching away from Gemini by applying provider defaults when model appears Gemini-specific.

## Why This Matters

The provider expansion is only successful if settings behavior is reliable in real usage:

- Provider selection must persist correctly.
- Model listing and connection testing must use the selected provider.
- Saved config must round-trip without silently falling back to Gemini assumptions.

Without this stabilization, the Phase 8 UI completion is incomplete in practice.

## Immediate Plan (Execution)

1. **Verify persistence path end-to-end**
   - Save OpenAI provider + model + key.
   - Reload settings and confirm values persist exactly.

2. **Verify provider-aware model listing path**
   - Confirm `/api/v2/ai/models` uses current provider settings after save.
   - Confirm expected fallback model behavior if provider model list fails.

3. **Run targeted tests/checks for touched behavior**
   - Backend API tests around AI settings and models route.
   - Frontend behavior checks around provider switch and save payload.

4. **Update planning/checklist artifacts with verified status**
   - Keep checklist aligned with completed verification work.
   - Keep this status doc updated until Phase 9 is complete.

## Definition of Done for Current Stabilization

This current stabilization effort is complete when:

- Provider no longer reverts unexpectedly.
- OpenAI (and other provider) save/reload behavior is consistent.
- Connection test and model fetch operate against the selected provider.
- Relevant tests pass for the changed behavior.
