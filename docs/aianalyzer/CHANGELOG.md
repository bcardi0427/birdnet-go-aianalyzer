# AI Analyzer Changelog

All notable AI Analyzer fork changes are documented here. The root `CHANGELOG.md` is kept aligned with the upstream BirdNET-Go project.

## [Unreleased]

### Features

- Added support for multiple LLM providers: Gemini, Anthropic, and OpenAI-compatible APIs (such as local LLMs or custom gateways).
- Added multi-provider settings to the AI Settings Page and secure encrypted configuration backend.
- Refactored frontend AI Analysis page and backend API handlers to support provider selection.
- Added UTM tracking parameter settings to the AI Settings Page and backend configurations to customize referral links.
- Updated AI analysis reports to stack external bird links in a cleaner layout.
- Added dynamic, dashboard-styled SVG initials for missing species thumbnail images in reports.
- Added `linux-arm64` binary target and release assets to the AI LXC Release GitHub Actions workflow for Raspberry Pi/arm64 support.
- Updated LXC upgrade script ([install-aianalyzer-lxc.sh](file:///F:/AntiGravity%20Sources/birdnet-go/scripts/install-aianalyzer-lxc.sh)) to automatically detect host architecture (`amd64` or `arm64`) and download the correct release binary.

### Fixes

- Fixed eBird URL generation by using correct species codes resolved from the BirdNET offline taxonomy database instead of hyphenated scientific names.
- Fixed TypeScript typecheck and compilation issues on the AI Settings Page, including making `utmParameters` a required field in AI Settings.
- Added missing untracked packages for embedded dependencies (including tflite/XNNPACK CGO/nocgo wrappers, ffmpeg path detection, and RTSP health integrations).

### Documentation

- Added the AI Analyzer Fork Handoff Guide under [AI_ASSISTANT_HANDOFF.md](file:///F:/AntiGravity%20Sources/birdnet-go/docs/aianalyzer/AI_ASSISTANT_HANDOFF.md).
- Added layout thumbnail link plan ([thumbnail_link_plan.md](file:///F:/AntiGravity%20Sources/birdnet-go/thumbnail_link_plan.md)) documenting the thumbnail image strategy.

## [1.0.0-beta.1] - 2026-05-20

### Features

- Added AI Analyzer daily report support with Gemini-backed report generation.
- Added AI report cache handling, including a logged-in-only cache bypass path to protect API token usage.
- Added AI settings support for enabling/disabling the feature, model selection, report days, cache hours, and prompt configuration.
- Added automatic encryption and decryption of API keys, passwords, and sensitive settings in `config.yaml` using AES-GCM with environment variable or key file resolver fallbacks.
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
