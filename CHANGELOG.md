# Changelog

All notable changes to this project will be documented in this file.

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

## [0.5.6] - 2026-06-05

### 🚀 Features

- _(birdnet)_ Implemented range filter auto-selection support for Perch v2 and BirdNET v3 models when enabled alongside BirdNET v2.4.

### 🐛 Bug Fixes

- _(birdnet)_ Fix range filter settings auto-selection bypass where startup configuration selections were ignored by subsequent ONNX initialization routines, defaulting back to legacy TFLite checklist and returning 6,522 species.
- _(birdnet)_ Automatically persist successfully initialized auto-selected range filter settings to `config.yaml` to prevent stale fallback.
- _(security)_ Resolved guest access authentication issue for AI endpoints by verifying access tokens manually inside `optionalAuthMiddleware` rather than calling auth middleware directly.
- _(weather)_ Enhanced weather report summaries by implementing a 5-minute server-side cache expiry to auto-refresh weather data.

## [0.5.5] - 2024-06-09

### 🚀 Features

- _(audio)_ Support for multiple RTSP streams
- _(birdnet)_ Range filter model is now selectable between latest and previous "legacy" version
- _(birdnet)_ Added "birdnet-go range print" command which lists all species included by range filter model
- _(birdnet)_ BirdNET overlap setting impacts now realtime process also

### 🐛 Bug Fixes

- _(privacy)_ Fix defaults for privacy and dog bark filters and fix incorrect setting names in default config.yaml
- _(privacy)_ Do not print dog bark detections if dog bark filter is not enabled
- _(privacy)_ Fix printf declaration for human detection confidence reporting
- _(audio)_ Yield to other goroutines in file_utils, policy_age, and policy_usage
- _(build)_ Linux/arm64 cross-compilation in docker build

### 🚜 Refactor

- _(audio)_ Refactor analysis and capture buffers to support multiple individual buffers for different audio sources
- _(audio)_ Move RTSP code to rtsp.go
- _(rtsp)_ Update default RTSP URLs to an empty slice
- _(rtsp)_ Update RTSP stream URLs example in config.yaml
- _(privacy)_ Make pricacy filter and dog bark filter source specific
- _(build)_ Move buildDate variable to main.go

### 🏗️ Building

- _(deps)_ Bump github.com/spf13/viper from 1.18.2 to 1.19.0
- _(deps)_ Bump github.com/shirou/gopsutil/v3 from 3.24.4 to 3.24.5

## [0.5.4] - 2024-06-01

### 🚀 Features

- _(audio)_ Disk usage based audio clip retention policy, enabled by default with 80% disk usage treshold
- _(conf)_ Privacy filter Confidence threshold setting
- _(conf)_ Dog bark filter Confidence threshold setting
- _(conf)_ Dog bark filter time to remember bark setting

### 🐛 Bug Fixes

- _(webui_) Fix Settings interface load error

### 🚜 Refactor

- _(conf)_ Refactor configuration package to improve settings handling, easier access to settings in code
- _(audio)_ Audio clip retention policy setting: none, age, usage
- _(audio)_ Age base audio clip retention policy accepts time in days, weeks, months and years instead of hours
- _(conf)_ Many settings renamed

### ⚙️ Miscellaneous Tasks

- Update go.mod with github.com/mitchellh/mapstructure v1.5.0

## [0.5.3] - 2024-05-21

### 🚀 Features

- _(birdweather)_ Added location fuzzing support for BirdWeather uploads, requires support for BirdWeather.com
- _(audio)_ Audio source device is now user configurable

### 🐛 Bug Fixes

- _(audio)_ Audio clip extraction fixed for occassional non-contiguous clips

### 🚜 Refactor

- _(conf)_ Move default config file from .go to .yaml and add proper default value initialization
- _(conf)_ Update audio export settings in updateconfig.go and realtime.go

### 🏗️ Building

- _(deps)_ Bump golang.org/x/text from 0.14.0 to 0.15.0
- _(deps)_ Bump golang.org/x/crypto from 0.22.0 to 0.23.0
- _(deps)_ Bump github.com/prometheus/client_golang

### ⚙️ Miscellaneous Tasks

- Update go version to 1.22.3
- Update golang version to 1.22.3
- Bump HTMX version from 1.9.11 to 1.9.12
- Bump daisyUI to 4.11.1
- Update custom.css to fix theme controller styles
- Update tailwindcss to v3.4.3
- Hide "Detections" column on smaller screens
- Update audio buffer initialization in realtime analysis
- Remove unused import and struct field in audiobuffer.go

## [0.5.2] - 2024-05-01

### 🐛 Bug Fixes

- File analysis restored
- Improve audio buffer write function time keeping
- _(birdweather)_ Improve handling of HTTP Responses in UploadSoundscape to prevent possible panics
- _(datastore)_ Refactor datastore Get, Delete and Save methods for efficient transaction and error handling
- _(tests)_ Refactor createDatabase function in interfaces_test.go for improved error handling
- _(datastore)_ Refactor GetClipsQualifyingForRemoval method in interfaces.go for improved input validation and error handling
- Refactor ClipCleanupMonitor function for improved error handling and logging
- _(birdweather)_ Fixed PCM to WAV encoding
- _(birdweather)_ Fixed PCM to WAV encoding
- _(birdweather)_ Increase HTTP timeout to 45 seconds
- _(utils)_ Do not report root user as missing from audio group
- _(audio)_ Fix default audio device reporting

### 💄 Enhancement

- _(audio)_ Print selected audio capture device on realtime mode startup
- _(startup)_ Enhance realtime mode startup message with system details to help troubleshooting

### 🚜 Refactor

- _(telemetry)_ Move Prometheus metrics to dedicated package and add pprof debug
- _(conf)_ Remove unused Context struct from internal/conf/context.go
- _(processor)_ Update range filter action to handle error when getting probable species

### 🏗️ Building

- _(deps)_ Bump golang.org/x/crypto from 0.21.0 to 0.22.0
- _(deps)_ Bump google.golang.org/protobuf from 1.32.0 to 1.33.0
- _(deps)_ Bump golang.org/x/net from 0.21.0 to 0.23.0
- _(go)_ Bump Go version from 1.21.6 to 1.22.2 in go.mod
- _(deps)_ Bump labstack echo version from 4.11.4 to 4.12.0
- _(deps)_ Bump gorm.io/gorm from 1.25.9 to 1.25.10
- _(deps)_ Bump github.com/gen2brain/malgo from 0.11.21 to 0.11.22

### ⚙️ Miscellaneous Tasks

- Fix linter errors
- Fix linter errors

### Github

- _(workflow)_ Add tensorflow dependencies to golangci-lint

## [0.5.1] - 2024-04-05

### 🐛 Bug Fixes

- _(birdnet)_ Make location filter threshold as configurable value under BirdNET node
- _(mqtt)_ Fix CodeRabbit magled code

### 🏗️ Building

- _(deps)_ Bump gorm.io/gorm from 1.25.8 to 1.25.9

## [0.5.0] - 2024-03-30

### 🚀 Features

- Privacy filter to discard audio clips with human vocals
- Save all BirdNET prediction results into table Results
- _(audio)_ Check user group membership on audio device open failure and print instructions for a fix
- _(docker)_ Added support for multiplatform build
- _(conf)_ New function to detect if running in container

### 🐛 Bug Fixes

- _(docker)_ Install ca-certificates package in container image
- _(capture)_ Set capture start to 5 seconds before detection instead of 4 seconds
- _(capture)_ Increase audio capture length from 9 to 12 seconds
- _(rtsp)_ Wait before restarting FFmpeg and update parent process on exit to prevent zombies

### 💄 Enhancement

- _(database)_ Switched sqlite journalling to MEMORY mode and added database optimize on closing
- _(workflow)_ Update GitHub Actions workflow to build and push Docker image to GHCR
- _(workflow)_ Update Docker actions versions
- _(workflow)_ Support multiplatform build with github actions
- _(docker)_ Add ffmpeg to container image
- _(labels)_ Add Greek (el) translations by @hoover67
- _(ui)_ Improve spectrogram generation to enable lazy loading of images
- _(make)_ Improve make file

### 🚜 Refactor

- Moved middleware code to new file
- Improved spectrogram generation
- Moved middleware code to new file
- _(database)_ Move save to common interface and change it to use transaction
- _(analyser)_ BirdNET Predict function and related types
- _(audio)_ Stopped audio device is now simply started again instead of full audio context restart
- _(audio)_ Disabled PulseAudio to prioritise audio capture to use ALSA on Linux
- _(audio)_ Set audio backend to fixed value based on OS
- _(config)_ Refactor RSTP config settings
- _(processor)_ Increase dog bark filter scope to 15 minutes and fix log messages
- _(rtsp)_ Improve FFmpeg process restarts and stopping on main process exit
- _(labels)_ Update makefile to zip labels.zip during build and have label files available in internal/birdnet/labels to make it easier to contribute language updates
- _(audio)_ Improve way start time of audio capture is calculated

### 📚 Documentation

- _(capture)_ Add documentation to audiobuffer.go
- Add git-cliff for changelog management
- _(changelog)_ Update git cliff config

### 🎨 Styling

- Remove old commented code
- _(docker)_ Removed commented out code

### 🏗️ Building

- _(deps)_ Add zip to build image during build
- _(deps)_ Bump gorm.io/driver/mysql from 1.5.4 to 1.5.6
- _(deps)_ Bump gorm.io/gorm from 1.25.7 to 1.25.8
- _(makefile)_ Update makefile
- _(makefile)_ Fix tensorflow lite lib install step
- _(makefile)_ Fix tflite install

### ⚙️ Miscellaneous Tasks

- _(assets)_ Upgrade htmx to 1.9.11

### Github

- _(workflow)_ Add windows build workflow
- _(workflow)_ Updated windows build workflow
- _(workflow)_ Add go lint workflow
- _(workflow)_ Remove obsole workflows
- _(workflow)_ Add build and attach to release workflow
- _(workflow)_ Update release-build.yml to trigger workflow on edited releases

## [0.3.0] - 2023-11-04

### 🚀 Features

- Added directory command
- Config file support
- Config file support

### 🐛 Bug Fixes

- Estimated time remaining print fixed
- Start and end time fix for stdout

<!-- generated by git-cliff -->
