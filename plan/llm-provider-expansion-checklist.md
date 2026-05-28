# LLM Provider Expansion Plan Checklist

- `[x]` **Phase 1: Add Provider-Neutral Settings**
  - `[x]` Add `Provider` and `BaseURL` to `conf.AISettings`
  - `[x]` Add default values
  - `[x]` Update YAML tag tests
  - `[x]` Update frontend `AISettings` TypeScript interface

- `[x]` **Phase 2: Update Validation and Normalization**
  - `[x]` Add provider validation against known provider map
  - `[x]` Normalize empty provider to `gemini`
  - `[x]` Normalize provider IDs (lowercase, trim space)
  - `[x]` Require model when AI is enabled
  - `[x]` Require API key or API key file conditionally
  - `[x]` Require or default base URL conditionally
  - `[x]` Replace Gemini-specific validation messages

- `[x]` **Phase 3: Introduce an LLM Provider Abstraction**
  - `[x]` Create `internal/ai/llm/` package
  - `[x]` Add `Provider` interface and request/response types
  - `[x]` Add factory `NewProvider` function

- `[x]` **Phase 4: Implement Provider Adapters**
  - `[x]` Gemini Adapter
  - `[x]` OpenAI Adapter
  - `[x]` OpenRouter Adapter
  - `[x]` OpenAI-Compatible Adapter
  - `[x]` Ollama Adapter
  - `[x]` Anthropic Adapter

- `[x]` **Phase 5: Refactor Report Generation**
  - `[x]` Rename Gemini-specific constants/functions
  - `[x]` Add helper functions (effectiveProvider, effectiveModel, etc.)
  - `[x]` Replace direct Gemini call with provider interface call
  - `[x]` Update fallback logs

- `[x]` **Phase 6: Update Cache Compatibility**
  - `[x]` Extend cache metadata to include new fields
  - `[x]` Update cache invalidation logic

- `[x]` **Phase 7: Update API Routes**
  - `[x]` Make `/api/v2/ai/models` provider-aware
  - `[x]` Resolve API key based on unified fields
  - `[x]` Return stable response shape or fallback model
  - `[x]` Update logs and error messages

- `[x]` **Phase 8: Update Frontend Settings UI**
  - `[x]` Add provider options dropdown
  - `[x]` Add base URL field
  - `[x]` Make API key help text provider-aware
  - `[x]` Update UI copy to be provider-neutral
  - `[x]` Update model placeholder based on provider

- `[~]` **Phase 9: Testing Plan**
  - `[ ]` Backend Config Tests
  - `[ ]` Backend API Tests
  - `[ ]` Report Service Tests
  - `[ ]` Provider Adapter Tests
  - `[ ]` Frontend Tests

- `[~]` **Documentation Plan**
  - `[~]` Update `docs/aianalyzer/README.md`
  - `[x]` Add configuration examples for each provider

## Progress Note (2026-05-21)

- Phase 1 has already been completed and is checked above.
- Current active work is beyond Phase 1 (stabilization + remaining testing/docs).
