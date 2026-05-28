# Runtime Version Markers (AI Provider Fix Tracking)

Last updated: 2026-05-21

This file documents where runtime/build markers are defined and where to verify them at runtime.

## 1) Backend runtime patch marker

- **Location in code:** `main.go`
- **Constant name:** `runtimePatchVersion`
- **Current value:**

```text
ai-provider-fix-2026-05-21-v2
```

- **Where it appears at runtime:** server startup log line from `main` module:

```text
INFO [main] BirdNET-Go starting ... runtime_patch=ai-provider-fix-2026-05-21-v2 ...
```

## 2) Frontend-visible version marker

- **UI location:** `Settings -> AI`
- **Component:** `frontend/src/lib/desktop/features/settings/pages/AISettingsPage.svelte`
- **Displayed label:** `Runtime build: <value>`
- **Data source:** `appState.version` from `/api/v2/app/config` response (`version` field)

## 3) API version source for frontend marker

- **API endpoint:** `GET /api/v2/app/config`
- **Backend handler:** `internal/api/v2/app.go` (`GetAppConfig`)
- **Response field:** `version`
- **Underlying source:** `c.Settings.Version` (populated in `main.go` from build-time `version` var)

## 4) Quick verification checklist

1. Start/restart server.
2. Confirm startup log includes:
   - `runtime_patch=ai-provider-fix-2026-05-21-v2`
3. Open **Settings -> AI**.
4. Confirm you see:
   - `Runtime build: ...`

## 5) How to update this document

When changing runtime markers:

1. Update the constant in `main.go`.
2. Update this file’s **Current value**.
3. Update `Last updated` date.
