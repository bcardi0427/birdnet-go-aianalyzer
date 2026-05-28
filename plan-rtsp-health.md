Phase 1 - Initial Understanding
- Objective: Diagnose RTSP health issues, ensure FFmpeg availability, and stabilize start script behavior in Windows environment.
- Focus areas: start.bat, FFmpeg path resolution, RTSP health checks in code paths (internal/api/v2/ai.go, settings.go), and server startup flow in birdnet-go.

Phase 2 - Design (plan to be used by next steps)
- Design goals:
  - Make FFmpeg callable from server startup with a configurable path; fall back gracefully if not found.
  - Add lightweight RTSP health probe that can be invoked on-demand and during startup.
  - Ensure start script (start.bat) sets PATH correctly and does not rely on non-existent test sources.
  - Preserve existing API key handling behavior in AI settings tests (no regression).
- Proposed changes:
  - Modify start.bat to locate FFmpeg in common install paths and add to PATH if found.
  - Instrument server startup to perform a quick RTSP health check using FFprobe/FFmpeg if available; log health status.
  - Introduce a optional --ffmpeg-path flag to birdnet-go serve to override PATH discovery.
  - Update tests around RTSP health to cover missing FFmpeg gracefully.

Phase 3 - Implementation Plan (high level steps)
- Step 1: Locate FFmpeg integration points in server startup code and settings parsing.
- Step 2: Implement a small helper that resolves FFmpeg path from env var or standard install locations.
- Step 3: Add RTSP health check function that returns status and health metrics.
- Step 4: Update start.bat to export FFmpeg path into environment and invoke health check on startup.
- Step 5: Run unit/integration tests focusing on RTSP and AI settings paths.

Phase 4 - Verification
- Validate that FFmpeg is available in PATH when starting the server in this environment.
- Confirm RTSP health is reported as healthy for a known-good stream (or gracefully degraded if not available).
- Ensure existing AI settings redaction/persistence tests pass.

Critical files to modify/include:
- start.bat
- internal/api/v2/settings.go
- internal/api/v2/ai.go
- birdnet-go startup path (wherever serve is invoked)

