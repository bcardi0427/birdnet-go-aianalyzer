# AI Analyzer Changelog

All notable AI Analyzer fork changes are documented here. The root `CHANGELOG.md` is kept aligned with the upstream BirdNET-Go project.

## [Unreleased] - 2026-05-20

### Features

- Added AI Analyzer daily report support with Gemini-backed report generation.
- Added AI report cache handling, including a logged-in-only cache bypass path to protect API token usage.
- Added AI settings support for enabling/disabling the feature, model selection, report days, cache hours, and prompt configuration.
- Added frontend AI Analysis and AI Settings pages.
- Added public cached AI report viewing while keeping refresh and bypass-cache actions protected.
- Added automatic hiding of the AI Analysis menu item when AI is disabled.
- Added Proxmox LXC install/upgrade flow for replacing the stock BirdNET-Go binary with the AI Analyzer build.
- Added GitHub Actions release workflow that publishes the `birdnet-go-linux-amd64` binary for easy LXC installs.
- Added public visitor logging to `logs/visitor.log`, including page path, status, IP, Cloudflare headers, referrer, user agent, authentication state, and tunnel metadata.
- Added admin-only Visitor Log dashboard under Settings for viewing recent visits, top pages, top IPs, countries, referrers, internal navigation, and AI report view counts.
- Added client-side SPA page-view tracking so internal navigation is recorded after the initial page load.
- Added session-based entry referrer tracking so later page views can still be attributed to the first external site that sent the visitor.

### Fixes

- Fixed AI report route behavior so guests can view cached reports without triggering fresh AI generation.
- Fixed AI report refresh behavior so only logged-in users can request fresh or cache-bypassed reports.
- Fixed missing frontend dependencies for markdown rendering and sanitizing AI report content.
- Fixed API wiring so AI report service is available from the v2 controller.
- Fixed missing AI report days configuration fields across backend and frontend settings.
- Fixed visitor stats so same-site navigation is separated from true external referrers.
- Fixed AI report view counting by recording client-side visits to `/ui/ai-analysis`.

### Documentation

- Added fork install documentation for `bcardi0427/birdnet-go-aianalyzer`.
- Added AI Analyzer documentation under `docs/aianalyzer/`.
- Added LXC install script documentation and curl-based install instructions.
