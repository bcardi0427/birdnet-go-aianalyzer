# Fix CodeRabbit Review Comments

This plan outlines the implementation of fixes for all 39 unaddressed CodeRabbit review comments identified in the latest scan of the workspace.

## User Review Required

- **Action PIN SHAs**: Pinning GitHub Action SHAs will require resolving the exact SHAs for the specified versions of `actions/checkout@v6`, `actions/setup-node@v3`, etc.
- **Environment Variables in Scripts**: Several Windows build/setup scripts (`build.ps1`, `setup_env.bat`, etc.) have hardcoded developer paths (like `C:\Users\Bcardi\src\tensorflow`). We will replace these with environment variable requirements (e.g., `TENSORFLOW_PATH`). Users running these scripts will need to define this env var going forward.
- **Artifact Deletion**: The `rollout-*.json` file will be deleted from git and added to `.gitignore`.

## Open Questions

- For the GitHub Actions SHA pinning, would you like me to fetch the latest SHAs automatically or just look up the specific SHAs for the tagged versions currently in use?
- For the environment variables in the Windows scripts, is `TENSORFLOW_PATH` acceptable as the standard name, and should I add a `.env.example` file to document it?

## Proposed Changes

### GitHub Workflows
#### [MODIFY] .github/workflows/aianalyzer-lxc-release.yml
- Pin GitHub actions to immutable commit SHAs instead of mutable tags (`@v6`, `@latest`, etc.).
- Add `persist-credentials: false` to the `actions/checkout` step to prevent token leakage to `.git/config`.

### Documentation
#### [MODIFY] CLAUDE.md
- Format markdown with Prettier to fix layout inconsistencies.
#### [MODIFY] internal/CLAUDE.md
- Replace developer-specific absolute paths with environment variables or examples.
#### [MODIFY] .agent/.skills/golang-pro.md
- Format markdown with Prettier to correct YAML frontmatter and code blocks.

### Frontend
#### [MODIFY] frontend/src/lib/desktop/features/settings/pages/AISettingsPage.svelte
- Relax the `requiresApiKey` logic so that `openai-compatible` providers do not strictly require an API key.
#### [MODIFY] frontend/src/lib/desktop/layouts/DesktopSidebar.svelte
- Replace the hardcoded `"Visitors"` string with the `t('settings.sections.visitors')` i18n function.
#### [MODIFY] frontend/static/messages/en.json
- Update AI-related UI copy to be provider-agnostic rather than exclusively referencing "Gemini".

### Internal AI & Configuration
#### [MODIFY] internal/ai/llm/defaults.go
- Add standard Go doc comments for all exported constants (e.g., `DefaultGeminiModel`).
#### [MODIFY] internal/ai/llm/openai_compatible.go
- Change constructor signatures to return the concrete type `*openAICompatibleProvider` instead of the `Provider` interface.
#### [MODIFY] internal/ai/service.go
- Add documentation for exported types and methods (`ReportService`, `ReportPayload`, `GetDailyReport`, `NewReportService`).
- Replace `fmt.Errorf` with the `internal/errors` package.
#### [MODIFY] internal/api/v2/ai.go
- Document all exported handlers (`GetAIModels`, `GetAIReport`, `isExplicitlyAuthenticated`, etc.).
- Refactor `UpdateAISettings` to extract the cyclomatically complex legacy/new provider key restoration logic into two helpers.
- Replace `fmt.Errorf` with `internal/errors`.
#### [MODIFY] internal/api/v2/settings.go
- Relax `ai` PATCH validation by removing the direct call to `validateAISection` to support partial updates properly.

### Internal Core Services
#### [MODIFY] internal/api/v2/media.go
- Prevent SVG/XML injection by escaping user-provided `initials` using `xml.EscapeText`.
#### [MODIFY] internal/api/v2/visitors.go
- Replace the hardcoded `"logs/visitor.log"` path with a configurable setting.
- Document exported functions.
#### [MODIFY] internal/conf/config.go
- Refactor the highly complex `MigrateAndSync` function to extract duplicated provider logic into small helpers, reducing cognitive complexity.
#### [MODIFY] internal/conf/secret_encryption.go
- Wrap returned errors with context using `fmt.Errorf("...: %w", err)` to simplify debugging.
#### [MODIFY] internal/conf/validate_ai_test.go
- Replace unsafe type assertions with the safe comma-ok pattern (`if v, ok := x.(Type); ok`).

### Internal FFmpeg & RTSP
#### [MODIFY] internal/ffmpeg/path.go
- Extract hardcoded Windows installation paths into named constants.
- Enhance `isExe` to validate against path traversal (`filepath.Clean`, `filepath.IsLocal`).
- Add tests for `GetFFmpegPath()`.
#### [MODIFY] internal/rtsp/health_integration.go
- Consolidate redundant wrapper functions into `ApplyBypassToHealthCheck` and `ShouldBypassHealthCheck`.

### Internal Secrets
#### [MODIFY] internal/secrets/secrets.go
- Trim inputs consistently before resolution so source classification aligns with what is actually resolved.

### Build & Setup Scripts
#### [MODIFY] build.ps1
- Replace hardcoded developer paths with environment variables (e.g., `$env:TENSORFLOW_PATH`).
#### [MODIFY] run_tests.ps1
- Replace hardcoded developer paths with environment variables.
#### [MODIFY] setup_env.bat
- Remove hardcoded paths; use `%CD%` and `%TENSORFLOW_PATH%`.
#### [MODIFY] setup_env.ps1
- Replace hardcoded paths with configurable environment checks.
#### [MODIFY] test_tax.go
- Add a proper error check before accessing the maps returned by `LoadTaxonomyData`.

### Repository Configuration
#### [DELETE] rollout-2026-05-16T08-12-52-019e30b4-252f-7a73-b3a1-df11c827e2a9.json
- Remove the tracked ephemeral JSON file from git.
#### [MODIFY] .gitignore
- Append `rollout-*.json` to ignore future rollout artifacts.
#### [MODIFY] .golangci.yaml
- Uncomment the `version: "2"` directive.

## Verification Plan
1. **Linting and Formatting**: Run `golangci-lint` and `prettier` to verify that formatting and complexity violations are resolved.
2. **Unit Tests**: Run `go test ./...` across the codebase, particularly targeting `internal/ffmpeg` and `internal/conf`.
3. **Build Scripts**: Run the modified `build.ps1` and `setup_env.ps1` locally to ensure they error cleanly when `TENSORFLOW_PATH` is missing, and succeed when it's provided.
