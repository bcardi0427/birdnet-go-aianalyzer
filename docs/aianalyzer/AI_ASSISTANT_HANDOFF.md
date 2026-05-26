# AI Analyzer Fork Handoff Guide

Use this document to orient another AI assistant or developer before working on this fork.

## Project Identity

This repository is `bcardi0427/birdnet-go-aianalyzer`, a fork of the upstream BirdNET-Go project at `tphakala/birdnet-go`.

The fork adds an AI Analyzer feature, visitor tracking, and Proxmox LXC install/upgrade tooling. The core BirdNET-Go project should remain as close to upstream as possible so future upstream merges do not become painful.

## Most Important Rule

Do not casually rewrite, reformat, rename, or reorganize upstream BirdNET-Go code.

When making fork-specific changes:

- Prefer files under `docs/aianalyzer/` for documentation.
- Keep AI Analyzer changes narrowly scoped.
- Avoid broad refactors.
- Avoid touching upstream docs unless the change is specifically about pointing users to the fork docs.
- Do not mix fork changelog entries into the root `CHANGELOG.md`.
- Use `docs/aianalyzer/CHANGELOG.md` for fork-specific changelog notes.

## Branches And Repository

Primary fork branch:

```text
aianalyzer/main
```

Remote repository:

```text
https://github.com/bcardi0427/birdnet-go-aianalyzer.git
```

Upstream repository:

```text
https://github.com/tphakala/birdnet-go.git
```

The intended sync model is:

```bash
git fetch upstream
git checkout main
git merge upstream/main
git push origin main
git checkout aianalyzer/main
git merge main
git push origin aianalyzer/main
```

## Fork Documentation Files

Fork docs live here:

```text
docs/aianalyzer/README.md
docs/aianalyzer/CHANGELOG.md
docs/aianalyzer/AI_ASSISTANT_HANDOFF.md
```

Root README may contain a short fork landing section, but detailed fork documentation belongs in `docs/aianalyzer/`.

## Installer Files

General fork installer:

```text
install-aianalyzer.sh
```

Proxmox helper-script LXC upgrade installer:

```text
scripts/install-aianalyzer-lxc.sh
```

The LXC installer is meant to run inside a BirdNET-Go LXC installed by the Proxmox VE helper script. That environment usually has:

```text
Service:          birdnet.service
Binary:           /usr/local/bin/birdnet-go
WorkingDirectory: /opt/birdnet/data
Logs:             /opt/birdnet/data/logs
Config:           /root/.config/birdnet-go/config.yaml
```

The LXC upgrade script backs up the current binary and replaces only `/usr/local/bin/birdnet-go`.

## Release Workflow

The LXC installer downloads the latest GitHub release asset:

```text
birdnet-go-linux-amd64
```

Release workflow:

```text
.github/workflows/aianalyzer-lxc-release.yml
```

If the installer receives a `404` for the binary, the release asset probably has not been generated yet or the workflow failed.

## AI Analyzer Feature Areas

Key backend areas:

```text
internal/ai/service.go
internal/api/v2/ai.go
internal/api/v2/api.go
internal/conf/config.go
internal/conf/defaults.go
internal/conf/validate_ai.go
internal/conf/secret_encryption.go
internal/conf/storage.go
```

Key frontend areas:

```text
frontend/src/lib/desktop/features/ai/AIAnalysisPage.svelte
frontend/src/lib/desktop/features/settings/pages/AISettingsPage.svelte
frontend/src/lib/utils/settingsApi.ts
frontend/src/App.svelte
```

The AI report page should be publicly viewable when a cached report exists, but actions that spend AI tokens must require authentication.

Protected/token-spending actions include:

- Refreshing the AI report.
- Bypassing the cache.
- Updating AI settings.

Guests should not be able to trigger fresh AI generation.

## Configuration Secrets Encryption

Sensitives settings (like AI API keys, MySQL/MQTT passwords, and session secrets) are automatically encrypted using AES-GCM when written to `config.yaml` and decrypted when read. 

Key resolution order:
1. `BIRDNET_CONFIG_ENCRYPTION_KEY` environment variable.
2. `config.encryption.key` file in the default config directory.
3. Automatically generated 32-byte key saved to `config.encryption.key` (with secure `0600` permissions) if neither of the above is found.

Any new sensitive settings should be added to the lists in `internal/conf/secret_encryption.go` so they are securely redacted/encrypted on storage write.

## Visitor Logging Feature Areas

Key backend file:

```text
internal/api/v2/visitors.go
```

Key frontend file:

```text
frontend/src/lib/desktop/features/settings/pages/VisitorLogSettingsPage.svelte
```

SPA page-view tracking is wired in:

```text
frontend/src/App.svelte
```

Visitor log output:

```text
/opt/birdnet/data/logs/visitor.log
```

The visitor log records public page visits, Cloudflare tunnel metadata, referrers, user agents, IPs, authentication state, and page paths.

The admin visitor dashboard is under Settings and should remain admin-only.

## Referrer Tracking Behavior

The app uses `sessionStorage`, not cookies, to preserve the first external referrer for a browser tab/session.

This avoids cookie banner requirements for this feature and lets later SPA visits, such as `/ui/ai-analysis`, still be attributed to the original external source.

Expected behavior:

- External referrers appear under `Top external referrers`.
- Same-site clicks appear under `Internal navigation`.
- AI report views count visits to `/ui/ai-analysis`.

## Cloudflare Tunnel Notes

The app may run behind Cloudflare Tunnel.

Important headers:

```text
CF-Connecting-IP
CF-IPCountry
CF-Ray
X-Forwarded-For
X-Forwarded-Proto
```

Do not assume all traffic comes from the local tunnel process. Cloudflare forwards the real visitor IP in headers when configured normally.

IPv6 visitor addresses are expected and valid.

## Authentication Expectations

Public users may view:

- Dashboard pages that upstream allows.
- Cached AI report page.

Logged-in/admin users may:

- Change settings.
- Refresh AI reports.
- Bypass AI report cache.
- View visitor log dashboard.

If authentication is changed, verify that public cached AI report viewing still works and token-spending actions still require login.

## Config Location In LXC

For the Proxmox helper-script LXC install, config is typically:

```text
/root/.config/birdnet-go/config.yaml
```

Restart command:

```bash
systemctl restart birdnet
```

Status/log commands:

```bash
systemctl status birdnet --no-pager
journalctl -u birdnet -n 80 --no-pager
tail -f /opt/birdnet/data/logs/visitor.log
```

## Build And Verification Notes

Frontend typecheck:

```bash
cd frontend
npm run typecheck
```

Linux release builds happen in GitHub Actions. Local Windows Go builds may fail because of CGO/TensorFlow/SQLite/audio dependencies. Do not treat those Windows CGO failures as proof that unrelated AI Analyzer code is broken.

Useful deployment test flow:

1. Push to `aianalyzer/main`.
2. Wait for GitHub Actions release workflow.
3. Re-run the LXC installer inside the BirdNET-Go LXC.
4. Restart/check `birdnet.service`.
5. Verify the app version in the sidebar.
6. Verify AI Analysis page behavior as guest and as logged-in admin.
7. Verify visitor log stats update after page navigation.

## Safe Change Checklist

Before committing:

- Confirm changed files are only the files intended for the current task.
- Do not stage unrelated local changes.
- Keep root upstream files untouched unless absolutely necessary.
- Put fork docs in `docs/aianalyzer/`.
- Put fork changelog entries in `docs/aianalyzer/CHANGELOG.md`.
- Run frontend typecheck for frontend changes.
- Let GitHub Actions validate the Linux binary release.

## Current User Intent

The user wants this fork to:

- Stay easy to merge with upstream BirdNET-Go.
- Install easily on a Proxmox helper-script LXC.
- Provide a public AI report page.
- Prevent guests from spending AI tokens.
- Provide admin-only visitor analytics.
- Keep fork documentation clear and separate from upstream documentation.
