# Changelog

All notable AI Analyzer fork changes will be documented in this file.

This changelog covers the `bcardi0427/birdnet-go-aianalyzer` fork. For inherited upstream BirdNET-Go history and credit, see the upstream [BirdNET-Go changelog](https://github.com/tphakala/birdnet-go/blob/main/CHANGELOG.md).

## [AI Analyzer 0.63] - 2026-06-08

### 🚀 Features

- _(ui)_ Added configurable bird thumbnail click destinations for detection details, eBird, Wikipedia, All About Birds, or no link, with separate thumbnail UTM parameters.
- _(ui)_ Extended configured thumbnail links to Detection Details hero images and Species Analytics species images across grid cards, table list view, and mobile cards.

### 🐛 Bug Fixes

- _(ui)_ Route non-bird thumbnail links to Wikipedia when eBird or All About Birds is selected, covering `t-` taxonomy codes such as Green Treefrog and known non-avian classes such as Dog.
- _(ui)_ Preserve existing internal detection detail links on thumbnails that already link to a detection instead of replacing them with external bird-site links.

### 📚 Documentation

- _(docs)_ Added wiki documentation files and updated the thumbnail link plan to document internal detection-link preservation.
- _(docs)_ Added `FIRESTORE_STATE_TRACKER_GUIDE.md` to `.gitignore`.

## [AI Analyzer 1.0.0-beta.1] - 2026-05-20

### 🚀 Features

- _(ai)_ Added AI Analyzer daily report support with Gemini-backed report generation.
- _(ai)_ Added AI report cache handling, including a logged-in-only cache bypass path to protect API token usage.
- _(ai)_ Added AI settings support for enabling/disabling the feature, model selection, report days, cache hours, and prompt configuration.
- _(security)_ Added automatic encryption and decryption of API keys, passwords, and sensitive settings in `config.yaml` using AES-GCM with environment variable or key file resolver fallbacks.
- _(ui)_ Added frontend AI Analysis and AI Settings pages.
- _(ai)_ Added public cached AI report viewing while keeping refresh and bypass-cache actions protected.
- _(ui)_ Added automatic hiding of the AI Analysis menu item when AI is disabled.
- _(install)_ Added Proxmox LXC install/upgrade flow for replacing the stock BirdNET-Go binary with the AI Analyzer build.
- _(build)_ Added GitHub Actions release workflow that publishes the `birdnet-go-linux-amd64` binary for easy LXC installs.
- _(visitors)_ Added public visitor logging to `logs/visitor.log`, including page path, status, IP, Cloudflare headers, referrer, user agent, authentication state, and tunnel metadata.
- _(visitors)_ Added admin-only Visitor Log dashboard under Settings for viewing recent visits, top pages, top IPs, countries, referrers, internal navigation, and AI report view counts.
- _(visitors)_ Added client-side SPA page-view tracking so internal navigation is recorded after the initial page load.
- _(visitors)_ Added session-based entry referrer tracking so later page views can still be attributed to the first external site that sent the visitor.

### 🐛 Bug Fixes

- _(ai)_ Fixed AI report route behavior so guests can view cached reports without triggering fresh AI generation.
- _(ai)_ Fixed AI report refresh behavior so only logged-in users can request fresh or cache-bypassed reports.
- _(frontend)_ Fixed missing frontend dependencies for markdown rendering and sanitizing AI report content.
- _(api)_ Fixed API wiring so AI report service is available from the v2 controller.
- _(settings)_ Fixed missing AI report days configuration fields across backend and frontend settings.
- _(visitors)_ Fixed visitor stats so same-site navigation is separated from true external referrers.
- _(visitors)_ Fixed AI report view counting by recording client-side visits to `/ui/ai-analysis`.

### 📚 Documentation

- _(docs)_ Added fork install documentation for `bcardi0427/birdnet-go-aianalyzer`.
- _(docs)_ Added AI Analyzer documentation under `docs/aianalyzer/`.
- _(docs)_ Added LXC install script documentation and curl-based install instructions.

## [AI Analyzer 2026-05-18]

### 🚀 Features

- _(ai)_ Added support for multiple LLM providers: Gemini, Anthropic, and OpenAI-compatible APIs such as local LLMs or custom gateways.
- _(ai)_ Added multi-provider settings to the AI Settings Page and secure encrypted configuration backend.
- _(ai)_ Refactored the frontend AI Analysis page and backend API handlers to support provider selection.
- _(ai)_ Added UTM tracking parameter settings to the AI Settings Page and backend configuration to customize referral links.
- _(ai)_ Updated AI analysis reports to stack external bird links in a cleaner layout.
- _(ui)_ Added dynamic dashboard-styled SVG initials for missing species thumbnail images in reports.
- _(build)_ Added `linux-amd64`, `linux-arm64`, `windows-amd64`, and `darwin-arm64` build targets to the release workflow, packaging binaries and required shared library dependencies into `.tar.gz` archives.
- _(install)_ Updated the LXC upgrade script to detect host CPU architecture (`amd64` or `arm64`) and fetch the correct raw binary.

### 🐛 Bug Fixes

- _(ai)_ Fixed eBird URL generation by using correct species codes resolved from the BirdNET offline taxonomy database instead of hyphenated scientific names.
- _(ai)_ Fixed TypeScript typecheck and compilation issues on the AI Settings Page, including making `utmParameters` a required field in AI Settings.
- _(build)_ Added missing untracked packages for embedded dependencies, including TensorFlow Lite/XNNPACK CGO and nocgo wrappers, ffmpeg path detection, and RTSP health integrations.
- _(build)_ Fixed TensorFlow Lite C library download failures on macOS build runners by preferring `curl -fsSL` and falling back to `wget` in `Taskfile.yml`.
- _(build)_ Fixed empty or corrupted ONNX Runtime caches breaking release builds by removing ONNX Runtime caching from `.github/actions/setup-onnxruntime/action.yml`.

### 📚 Documentation

- _(docs)_ Added the AI Analyzer Fork Handoff Guide under `docs/aianalyzer/AI_ASSISTANT_HANDOFF.md`.
- _(docs)_ Added the thumbnail link plan documenting the thumbnail image strategy.
