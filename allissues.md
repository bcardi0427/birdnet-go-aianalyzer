# Upstream Issues Analysis & Classification

This report classifies the public open issues from [tphakala/birdnet-go/issues](https://github.com/tphakala/birdnet-go/issues) and cross-references them with our local codebase.

## Summary Statistics
- **Total Open Issues Checked**: 77
- **High Priority / Good to Fix**: 30
- **Easy to Fix / Direct Layout**: 12
- **Medium Feasibility**: 51
- **Hard / Massive Structural Changes**: 14

---

## 1. High Priority / Good to Fix
These are bugs, crashes, or regressions that affect core functionality, correctness, or robustness in our version.

| Issue | Title | Labels | Key Analysis / Code Location |
| :--- | :--- | :--- | :--- |
| [#3250](https://github.com/tphakala/birdnet-go/issues/3250) | bug: Species range filter not working | *bug, confirmed* | `internal/analysis/range_filter.go` |
| [#3243](https://github.com/tphakala/birdnet-go/issues/3243) | bug: After update to nightly-20260523 "New Species" star shows on all detections | *bug, pkg:api* | General |
| [#3225](https://github.com/tphakala/birdnet-go/issues/3225) | bug: weak detection results | *bug, pkg:birdnet, has-dump* | `internal/classifier/` (model inference) |
| [#3221](https://github.com/tphakala/birdnet-go/issues/3221) | Webhook push provider receives empty/missing detection metadata for rules-engine (alert) notifications | *bug, pkg:notification* | General |
| [#3126](https://github.com/tphakala/birdnet-go/issues/3126) | bug: perch v2 checkbox disappeared from audio interface | *bug, pkg:api* | `frontend/src/lib/...` |
| [#3121](https://github.com/tphakala/birdnet-go/issues/3121) | bug: Perch detection considered new species but same species already exists | *bug, pkg:analysis, translation, has-dump* | General |
| [#3103](https://github.com/tphakala/birdnet-go/issues/3103) | bug: After entering weather provider's api, the api is lost even after saving | *bug, pkg:weather* | `internal/api/v2/settings.go` |
| [#3000](https://github.com/tphakala/birdnet-go/issues/3000) | bug: Let's Encrypt (autotls) configuration not working | *bug, has-dump* | General |
| [#2837](https://github.com/tphakala/birdnet-go/issues/2837) | bug: Systematic Old World taxon misidentification in North American deployments — 5 species requiring location-aware code remapping | *bug, pkg:analysis, pkg:birdnet* | General |
| [#2679](https://github.com/tphakala/birdnet-go/issues/2679) | Web UI audio source change corrupts config and crashes BirdNET-Go on restart | *None* | General |
| [#2599](https://github.com/tphakala/birdnet-go/issues/2599) | Title: printSystemDetails nil pointer dereference on macOS ARM64 under launchd | *None* | `internal/analysis/startup.go` |
| [#2389](https://github.com/tphakala/birdnet-go/issues/2389) | api/v2/analytics/species/summary seems to be missing recent detections | *None* | General |
| [#2206](https://github.com/tphakala/birdnet-go/issues/2206) | improvement: FFmpeg process killed by OOM in memory-constrained containers | *enhancement* | Memory limits in FFmpeg process |
| [#2204](https://github.com/tphakala/birdnet-go/issues/2204) | improvement: Add graceful handling for disk-full conditions | *enhancement* | Disk monitoring in database service |
| [#2202](https://github.com/tphakala/birdnet-go/issues/2202) | improvement: Spectrogram pre-render queue too small under load | *enhancement* | General |
| [#1905](https://github.com/tphakala/birdnet-go/issues/1905) | Implement proper species sorting via label JOIN | *enhancement* | General |
| [#1904](https://github.com/tphakala/birdnet-go/issues/1904) | Implement reconciliation job for dual-write dirty IDs | *enhancement* | General |
| [#1870](https://github.com/tphakala/birdnet-go/issues/1870) | RTSP not working in Nightly Build nightly-20260118 | *bug* | General |
| [#1842](https://github.com/tphakala/birdnet-go/issues/1842) | Migrate remaining Note field consumers to detection.Result | *enhancement* | General |
| [#1813](https://github.com/tphakala/birdnet-go/issues/1813) | sox and ffmpeg versions etc. | *bug* | General |
| [#1758](https://github.com/tphakala/birdnet-go/issues/1758) | Analysis Processor unable to find Custom Script for Species Custom Action | *bug* | General |
| [#1720](https://github.com/tphakala/birdnet-go/issues/1720) | New BirdNET-Go install errors | *bug* | General |
| [#1462](https://github.com/tphakala/birdnet-go/issues/1462) | Better host/url setting when proxied | *enhancement* | General |
| [#876](https://github.com/tphakala/birdnet-go/issues/876) | RFC: Complete Audio Package Rewrite - Modular Architecture with Multi-Source Support | *enhancement* | `internal/myaudio` rewrite |
| [#875](https://github.com/tphakala/birdnet-go/issues/875) | RFC: Non-Blocking Database Migration Strategy | *enhancement, help wanted* | General |
| [#874](https://github.com/tphakala/birdnet-go/issues/874) | RFC: Optimized Database Schema Design | *enhancement, help wanted* | General |
| [#865](https://github.com/tphakala/birdnet-go/issues/865) | Performance: Implement memory pooling for high-frequency buffer allocations in myaudio package | *enhancement* | General |
| [#838](https://github.com/tphakala/birdnet-go/issues/838) | Feature: implement batch reporting to Sentry in TelemetryWorker | *enhancement* | General |
| [#827](https://github.com/tphakala/birdnet-go/issues/827) | Feature: Add error deduplication cache to reduce telemetry volume | *enhancement* | General |
| [#822](https://github.com/tphakala/birdnet-go/issues/822) | Architecture: Separate audio data from Results struct to reduce memory coupling | *enhancement* | General |

---

## 2. Feasibility Check
Classification of all issues by how easily they can be implemented or fixed within our current layout versus requiring massive structural changes.

### A. Easy / Direct Fixes
Can be implemented with simple localization updates, direct UI fixes, or minor Go logic additions.

| Issue | Title | Scope |
| :--- | :--- | :--- |
| [#3161](https://github.com/tphakala/birdnet-go/issues/3161) | feat: separate secrets from config.yaml into dedicated secret store | Secret separation enhancement |
| [#3103](https://github.com/tphakala/birdnet-go/issues/3103) | bug: After entering weather provider's api, the api is lost even after saving | Go Weather settings trigger |
| [#2847](https://github.com/tphakala/birdnet-go/issues/2847) | feat: request wakelock on live audio tab | Minor Bug / Enhancement |
| [#2698](https://github.com/tphakala/birdnet-go/issues/2698) | Playback controls in search results table are missing for screen width >= 768 px | Minor Bug / Enhancement |
| [#2682](https://github.com/tphakala/birdnet-go/issues/2682) | Minor feature request: Sum of daily species seen | Minor Bug / Enhancement |
| [#2599](https://github.com/tphakala/birdnet-go/issues/2599) | Title: printSystemDetails nil pointer dereference on macOS ARM64 under launchd | Go Panic recovery in `startup.go` |
| [#2237](https://github.com/tphakala/birdnet-go/issues/2237) | Allow the name in the sidebar to reflect the node name | Frontend Svelte change (Sidebar node name) |
| [#1986](https://github.com/tphakala/birdnet-go/issues/1986) | Feature Request: Ability to exclude specific species from saved Clip Recordings while still tracking detections | Minor Bug / Enhancement |
| [#1948](https://github.com/tphakala/birdnet-go/issues/1948) | 15 Hungarian translation | Minor Bug / Enhancement |
| [#1672](https://github.com/tphakala/birdnet-go/issues/1672) | "latest" docker tag not being maintained | Minor Bug / Enhancement |
| [#1254](https://github.com/tphakala/birdnet-go/issues/1254) | Feature Request - Link to Educational Sites for Detected Species | Minor Bug / Enhancement |
| [#1214](https://github.com/tphakala/birdnet-go/issues/1214) | Requesting to Add the Species: Malleefowl | Minor Bug / Enhancement |

### B. Medium Feasibility
Requires moderate updates to Go logic, new database queries, API endpoints, or more extensive frontend components, but still within our current architectural boundary.

| Issue | Title | Scope |
| :--- | :--- | :--- |
| [#3250](https://github.com/tphakala/birdnet-go/issues/3250) | bug: Species range filter not working | Subsystem Update |
| [#3243](https://github.com/tphakala/birdnet-go/issues/3243) | bug: After update to nightly-20260523 "New Species" star shows on all detections | Subsystem Update |
| [#3225](https://github.com/tphakala/birdnet-go/issues/3225) | bug: weak detection results | Subsystem Update |
| [#3221](https://github.com/tphakala/birdnet-go/issues/3221) | Webhook push provider receives empty/missing detection metadata for rules-engine (alert) notifications | Subsystem Update |
| [#3151](https://github.com/tphakala/birdnet-go/issues/3151) | bug: Species filter drops to 0 and possibilities don't correspond to location | Subsystem Update |
| [#3126](https://github.com/tphakala/birdnet-go/issues/3126) | bug: perch v2 checkbox disappeared from audio interface | Subsystem Update |
| [#3121](https://github.com/tphakala/birdnet-go/issues/3121) | bug: Perch detection considered new species but same species already exists | Subsystem Update |
| [#3092](https://github.com/tphakala/birdnet-go/issues/3092) | feat: Migration Explorer - interactive species distribution heatmap | Subsystem Update |
| [#3000](https://github.com/tphakala/birdnet-go/issues/3000) | bug: Let's Encrypt (autotls) configuration not working | Subsystem Update |
| [#2951](https://github.com/tphakala/birdnet-go/issues/2951) | feat: Explainable failed detections | Subsystem Update |
| [#2947](https://github.com/tphakala/birdnet-go/issues/2947) | feat: Provide a way to listen to good examples of a bird's song, to compare to | Subsystem Update |
| [#2890](https://github.com/tphakala/birdnet-go/issues/2890) | feat: Seperate backend vs frontend localization to support multiple client languages | Subsystem Update |
| [#2838](https://github.com/tphakala/birdnet-go/issues/2838) | feat: Select several detections for deletion | Subsystem Update |
| [#2837](https://github.com/tphakala/birdnet-go/issues/2837) | bug: Systematic Old World taxon misidentification in North American deployments — 5 species requiring location-aware code remapping | Taxonomy remapping / validation logic |
| [#2820](https://github.com/tphakala/birdnet-go/issues/2820) | feat: Configuration Backup | Configuration backup/restore endpoints |
| [#2813](https://github.com/tphakala/birdnet-go/issues/2813) | feat: Include details about which RTSP stream for detections | Subsystem Update |
| [#2759](https://github.com/tphakala/birdnet-go/issues/2759) | feat: Clarify what Sensitivity, Threshold, and Overlap settings do | Subsystem Update |
| [#2741](https://github.com/tphakala/birdnet-go/issues/2741) | Feature Request: Exclusion lists for both False Positive Filter, and Learned birds | Subsystem Update |
| [#2702](https://github.com/tphakala/birdnet-go/issues/2702) | Live spectrogram waterfall is blank on Safari/WebKit; propose server-side FFT streaming for all browsers plus x- and y-axes | Safari compatibility / streaming spectrogram |
| [#2700](https://github.com/tphakala/birdnet-go/issues/2700) | OIDC Authentik Error: no provider for openid-connect exists | Subsystem Update |
| [#2697](https://github.com/tphakala/birdnet-go/issues/2697) | More sorting/searching/filtering options | Subsystem Update |
| [#2691](https://github.com/tphakala/birdnet-go/issues/2691) | Changing date range on analytics page doesn't change stats | Subsystem Update |
| [#2679](https://github.com/tphakala/birdnet-go/issues/2679) | Web UI audio source change corrupts config and crashes BirdNET-Go on restart | Subsystem Update |
| [#2664](https://github.com/tphakala/birdnet-go/issues/2664) | Add volume control | Subsystem Update |
| [#2661](https://github.com/tphakala/birdnet-go/issues/2661) | Add basic privacy filter observability metrics for human-voice hits and last activation time | Subsystem Update |
| [#2638](https://github.com/tphakala/birdnet-go/issues/2638) | rtsp Stream from Reolink P850 Camera cannot be connected | Subsystem Update |
| [#2592](https://github.com/tphakala/birdnet-go/issues/2592) | Add TLS certificate verification option for RTSPS audio streams | Subsystem Update |
| [#2539](https://github.com/tphakala/birdnet-go/issues/2539) | Species code mismatch: hergul should be amhgul1 (American Herring Gull taxonomic split) | Subsystem Update |
| [#2495](https://github.com/tphakala/birdnet-go/issues/2495) | Add maximum clips stored option + live logs / spectrograms | Subsystem Update |
| [#2491](https://github.com/tphakala/birdnet-go/issues/2491) | RFE: Ability to choose source for live spectrogram? | Subsystem Update |
| [#2389](https://github.com/tphakala/birdnet-go/issues/2389) | api/v2/analytics/species/summary seems to be missing recent detections | Subsystem Update |
| [#2345](https://github.com/tphakala/birdnet-go/issues/2345) | Add a nighttime filter to complement the daylight filter | Subsystem Update |
| [#1905](https://github.com/tphakala/birdnet-go/issues/1905) | Implement proper species sorting via label JOIN | Subsystem Update |
| [#1904](https://github.com/tphakala/birdnet-go/issues/1904) | Implement reconciliation job for dual-write dirty IDs | Subsystem Update |
| [#1895](https://github.com/tphakala/birdnet-go/issues/1895) | Feature Request: API Endpoint or Internal Scheduler for RTSP Audio Detection Control | Subsystem Update |
| [#1870](https://github.com/tphakala/birdnet-go/issues/1870) | RTSP not working in Nightly Build nightly-20260118 | Subsystem Update |
| [#1842](https://github.com/tphakala/birdnet-go/issues/1842) | Migrate remaining Note field consumers to detection.Result | Subsystem Update |
| [#1813](https://github.com/tphakala/birdnet-go/issues/1813) | sox and ffmpeg versions etc. | Subsystem Update |
| [#1758](https://github.com/tphakala/birdnet-go/issues/1758) | Analysis Processor unable to find Custom Script for Species Custom Action | Subsystem Update |
| [#1724](https://github.com/tphakala/birdnet-go/issues/1724) | Use both channels of stereo mic? | Subsystem Update |
| [#1720](https://github.com/tphakala/birdnet-go/issues/1720) | New BirdNET-Go install errors | Subsystem Update |
| [#1703](https://github.com/tphakala/birdnet-go/issues/1703) | [Feature Request] make sonagram "look & feel" configurable | Subsystem Update |
| [#1673](https://github.com/tphakala/birdnet-go/issues/1673) | [Feature Request] Reduce webui data transfer | Subsystem Update |
| [#1462](https://github.com/tphakala/birdnet-go/issues/1462) | Better host/url setting when proxied | Subsystem Update |
| [#1425](https://github.com/tphakala/birdnet-go/issues/1425) | Feature request : Sirens detection location based | Subsystem Update |
| [#1283](https://github.com/tphakala/birdnet-go/issues/1283) | Make sorting possible | Subsystem Update |
| [#1255](https://github.com/tphakala/birdnet-go/issues/1255) | Feature Request - More Detailed Spectrograms | Subsystem Update |
| [#1183](https://github.com/tphakala/birdnet-go/issues/1183) | Analytics - More trends graphs | Subsystem Update |
| [#1155](https://github.com/tphakala/birdnet-go/issues/1155) | Add information whether birdnet-go will actually learn from my reviews | Subsystem Update |
| [#838](https://github.com/tphakala/birdnet-go/issues/838) | Feature: implement batch reporting to Sentry in TelemetryWorker | Subsystem Update |
| [#827](https://github.com/tphakala/birdnet-go/issues/827) | Feature: Add error deduplication cache to reduce telemetry volume | Subsystem Update |

### C. Hard / Massive Structural Changes
Requires a major architecture overhaul, rewriting core components (e.g. database migrations, audio pipelines), or implementing hardware-specific optimizations.

| Issue | Title | Scope |
| :--- | :--- | :--- |
| [#3239](https://github.com/tphakala/birdnet-go/issues/3239) | feat: OpenVINO Execution Provider support for Intel iGPU offloading Perch v2 | Intel iGPU OpenVINO support |
| [#2206](https://github.com/tphakala/birdnet-go/issues/2206) | improvement: FFmpeg process killed by OOM in memory-constrained containers | Major Architecture / RFC |
| [#2204](https://github.com/tphakala/birdnet-go/issues/2204) | improvement: Add graceful handling for disk-full conditions | Major Architecture / RFC |
| [#2202](https://github.com/tphakala/birdnet-go/issues/2202) | improvement: Spectrogram pre-render queue too small under load | Major Architecture / RFC |
| [#2161](https://github.com/tphakala/birdnet-go/issues/2161) | feat(app): eliminate package-level globals and remove LegacyService wrapper | Eliminate globals / LegacyService refactoring |
| [#2160](https://github.com/tphakala/birdnet-go/issues/2160) | feat(app): extract BirdNET into Analyzer service with source routing | Major Architecture / RFC |
| [#2159](https://github.com/tphakala/birdnet-go/issues/2159) | feat(app): extract audio capture into AudioPipelineService | Major Architecture / RFC |
| [#2158](https://github.com/tphakala/birdnet-go/issues/2158) | feat(app): extract API server into APIServerService with TierNetwork shutdown | Major Architecture / RFC |
| [#2157](https://github.com/tphakala/birdnet-go/issues/2157) | feat(app): extract database into DatabaseService with TierCore shutdown | Major Architecture / RFC |
| [#876](https://github.com/tphakala/birdnet-go/issues/876) | RFC: Complete Audio Package Rewrite - Modular Architecture with Multi-Source Support | Audio capture rewrite |
| [#875](https://github.com/tphakala/birdnet-go/issues/875) | RFC: Non-Blocking Database Migration Strategy | Major Architecture / RFC |
| [#874](https://github.com/tphakala/birdnet-go/issues/874) | RFC: Optimized Database Schema Design | Major Architecture / RFC |
| [#865](https://github.com/tphakala/birdnet-go/issues/865) | Performance: Implement memory pooling for high-frequency buffer allocations in myaudio package | Major Architecture / RFC |
| [#822](https://github.com/tphakala/birdnet-go/issues/822) | Architecture: Separate audio data from Results struct to reduce memory coupling | Major Architecture / RFC |