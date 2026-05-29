# Upstream Issues Fix Plan

This implementation plan covers the top 3 easiest and most beneficial issues selected from the upstream BirdNET-Go issues list, to be resolved in our local codebase.

## User Review Required

> [!IMPORTANT]
> - **Weather Reload Mechanism**: This plan introduces a new `reconfigure_weather` action to the `ControlMonitor` that terminates the running polling loop of the weather service and restarts it with updated parameters from the config without requiring a full application restart.
> - **i18n & Type Generation**: We will add a new key `"reconfiguringWeather"` to `en.json` and then run the frontend type generator `npm run generate:i18n-types` to align the TypeScript bindings.
> - **macOS host.Info Panic Recovery**: We wrap `host.Info()` in a panic recovery block to handle launchd/sandbox environments safely.

## Proposed Changes

### Startup Diagnostics
#### [MODIFY] [startup.go](file:///F:/AntiGravity%20Sources/birdnet-go/internal/analysis/startup.go)
- Add a guard `if settings == nil { return }` at the start of `PrintSystemDetails`.
- Wrap `host.Info()` in a panic recovery block to prevent startup crashes when executing inside launchd on macOS.

---

### Weather Service Settings Reload
#### [MODIFY] [message_keys.go](file:///F:/AntiGravity%20Sources/birdnet-go/internal/notification/message_keys.go)
- Add `MsgSettingsReconfiguringWeather = "notifications.content.settings.reconfiguringWeather"`.

#### [MODIFY] [en.json](file:///F:/AntiGravity%20Sources/birdnet-go/frontend/static/messages/en.json)
- Add `"reconfiguringWeather": "Updating weather service settings..."` to the settings notifications block.

#### [MODIFY] [settings.go](file:///F:/AntiGravity%20Sources/birdnet-go/internal/api/v2/settings.go)
- Implement `weatherSettingsChanged` comparison helper.
- Add `{"Weather", "reconfigure_weather", weatherSettingsChanged, "Reconfiguring weather service...", notification.MsgSettingsReconfiguringWeather, "info", toastDurationMedium}` to `settingsChangeChecks`.

#### [MODIFY] [control_monitor.go](file:///F:/AntiGravity%20Sources/birdnet-go/internal/analysis/control_monitor.go)
- Add `reconfigureWeatherFn func()` field to `ControlMonitor` and accept it in `NewControlMonitor`.
- Map the `"reconfigure_weather"` signal in `handleControlSignal` to call `handleReconfigureWeather()`.

#### [MODIFY] [audio_pipeline_service.go](file:///F:/AntiGravity%20Sources/birdnet-go/internal/analysis/audio_pipeline_service.go)
- Add `weatherMu`, `weatherService`, and `weatherStop` to the `AudioPipelineService` struct to manage its lifecycle.
- Update `startWeatherPolling` to save the stop channel and service reference.
- Implement `reconfigureWeather` method which gracefully closes the active stop channel and reinstantiates the weather service.
- Update `Stop()` to cleanly stop the weather service.

---

### Sidebar Node Name Customization
#### [MODIFY] [DesktopSidebar.svelte](file:///F:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/desktop/layouts/DesktopSidebar.svelte)
- Import `mainSettings` store: `import { mainSettings } from '$lib/stores/settings';`
- Bind the header logo text to `{$mainSettings?.name || 'BirdNET-Go'}`.

## Verification Plan

### Automated Tests
- Run `go test ./internal/analysis/...` and `go test ./internal/api/v2/...` to verify that Go changes pass compilation and tests.
- Run frontend type generation: `npm run --prefix frontend generate:i18n-types` and check that type validation passes: `npm run --prefix frontend i18n:validate`.
- Run frontend unit tests: `npm run --prefix frontend test`.

### Manual Verification
- Launch the application and update the Weather API key or provider options via the UI. Verify that the weather service polling loop restarts with the new configuration, sending a toast confirmation.
- Modify the node name in Settings -> Main Settings. Verify that the navigation sidebar header dynamically reflects the new name.
