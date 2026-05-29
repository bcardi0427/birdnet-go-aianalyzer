# Parallel Implementation Plan: Upstream Issues

This plan describes three independent front-end changes designed to be implemented by three separate agents in parallel. All tasks are isolated to specific files to avoid git merge conflicts.

## User Review Required

> [!IMPORTANT]
> - **Search Playback overlay component (`MobileAudioPlayer.svelte`)**: The mobile audio overlay is currently restricted to mobile view (`md:hidden`). We will make it responsive so that on desktop screens (>= 768px), it displays as a beautiful centered modal dialog instead of a bottom sheet. This allows us to reuse the same audio player component seamlessly.
> - **Screen Wake Lock API**: Wake locks will be active reactively when HLS streaming is active on the Live Audio page. If the user pauses/stops streaming or closes the tab, the wake lock is released.
> - **i18n & Type Generation**: Adding new translation keys to `en.json` requires running `npm run generate:i18n-types` under the `frontend` folder to update i18n type definitions.

---

## Proposed Changes

### Component 1: Search Playback Controls (#2698)
**Assigned to Agent 1**

Enable the playback action button in the search results table on desktop view.

#### [MODIFY] [MobileAudioPlayer.svelte](file:///f:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/desktop/components/media/MobileAudioPlayer.svelte)
- Update the layout styles so the component renders as a centered dialog modal on screen widths `>= 768px` instead of being hidden.
- Add click-outside/backdrop closing support.
- Implement the style changes:
  - Container class:
    - Change `class="fixed inset-0 z-50 bg-black/50 flex items-end md:hidden"`
    - To `class="fixed inset-0 z-50 bg-black/50 flex items-end md:items-center md:justify-center"` and add an `onclick={handleClose}` handler.
  - Dialog body wrapper:
    - Change `class="w-full rounded-t-3xl shadow-2xl relative overflow-hidden bg-[var(--color-base-100)]"`
    - To `class="w-full rounded-t-3xl md:rounded-2xl shadow-2xl relative overflow-hidden bg-[var(--color-base-100)] md:max-w-[480px] md:m-4"` and add `onclick={e => e.stopPropagation()}` to prevent backdrop click closing when clicking inside the player.

#### [MODIFY] [Search.svelte](file:///f:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/desktop/views/Search.svelte)
- Replace the placeholder `// TODO: Implement audio playback function` in the actions column button (around line 939) with a call to `openMobilePlayer(result)`.
- Remove the `<div class="md:hidden">` wrapper around `<MobileAudioPlayer>` rendering block (around line 1221) to allow the player modal to be visible on desktop screens.

---

### Component 2: Screen Wake Lock on Live Audio tab (#2847)
**Assigned to Agent 2**

Request a screen wake lock when live audio streaming is active to prevent screens from sleeping during observation.

#### [MODIFY] [LiveStreamPage.svelte](file:///f:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/desktop/features/live-stream/pages/LiveStreamPage.svelte)
- Manage the screen wake lock reactively based on streaming status.
- Implementation details:
  - Add state: `let wakeLock = $state<any>(null);`
  - Implement `requestWakeLock()` to request lock with `'screen'` if supported by the browser:
    ```typescript
    async function requestWakeLock() {
      if (!('wakeLock' in navigator)) return;
      try {
        if (wakeLock) await wakeLock.release().catch(() => {});
        wakeLock = await navigator.wakeLock.request('screen');
        logger.info('Screen Wake Lock acquired successfully');
      } catch (err) {
        logger.warn('Failed to acquire screen wake lock:', err);
      }
    }
    ```
  - Implement `releaseWakeLock()` to release the lock:
    ```typescript
    async function releaseWakeLock() {
      if (wakeLock) {
        await wakeLock.release().catch(() => {});
        wakeLock = null;
        logger.info('Screen Wake Lock released');
      }
    }
    ```
  - Add an effect block to request the wake lock when `isStreaming` is active, handle page visibility changes, and release on stream stop or page unmount:
    ```typescript
    $effect(() => {
      if (isStreaming) {
        requestWakeLock();
        
        const handleVisibilityChange = () => {
          if (document.visibilityState === 'visible' && isStreaming) {
            requestWakeLock();
          }
        };

        document.addEventListener('visibilitychange', handleVisibilityChange);
        return () => {
          document.removeEventListener('visibilitychange', handleVisibilityChange);
          releaseWakeLock();
        };
      } else {
        releaseWakeLock();
      }
    });
    ```

---

### Component 3: Sum of Daily Species Seen (#2682)
**Assigned to Agent 3**

Compute and display the total unique species seen on the dashboard's daily activity heatmap.

#### [MODIFY] [en.json](file:///f:/AntiGravity Sources/birdnet-go/frontend/static/messages/en.json)
- Add localized species count templates under `"dailySummary"` section:
  ```json
  "totalSpeciesZero": "No species seen",
  "totalSpeciesOne": "1 species seen",
  "totalSpeciesOther": "{count} species seen",
  ```

#### [MODIFY] [DailySummaryCard.svelte](file:///f:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/desktop/features/dashboard/components/DailySummaryCard.svelte)
- Implement a helper to retrieve the correct pluralized string:
  ```typescript
  function getSpeciesSeenLabel(count: number): string {
    if (count === 0) {
      return t('dashboard.dailySummary.totalSpeciesZero');
    }
    if (count === 1) {
      return t('dashboard.dailySummary.totalSpeciesOne');
    }
    return t('dashboard.dailySummary.totalSpeciesOther', { count });
  }
  ```
- Modify the subtitle display line within the `'loaded'` loading phase branch (around line 867) to append the total species count:
  ```svelte
  <p class="text-sm text-[var(--color-base-content)]/60">
    {t('dashboard.dailySummary.subtitle')} • {getSpeciesSeenLabel(data.length)}
  </p>
  ```

---

## Verification Plan

### Automated Tests
Execute the following verification scripts under the `frontend` directory:
- Type generation: `npm run --prefix frontend generate:i18n-types`
- Translation validation: `npm run --prefix frontend i18n:validate`
- Type checking: `npm run --prefix frontend typecheck`
- Code formatting & Linting: `npm run --prefix frontend check`
- Unit testing: `npm run --prefix frontend test:ci`

### Manual Verification
- **Search Playback**: Conduct a search, click the play button in a results row on a desktop viewport, and confirm the audio player modal displays in the center of the screen. Verify clicking the gray overlay background closes the modal.
- **Screen Wake Lock**: Navigate to the Live Stream page, click play to start streaming, and verify (via browser console debugging or wake-lock diagnostics) that a screen wake lock is successfully requested. Stop streaming or change tabs and check that the wake lock is released.
- **Daily Species Seen**: Open the dashboard view. The heatmap subtitle should now display the total count of species identified (e.g. `Species detections by hour • 14 species seen`). Check navigation between days to ensure the count updates correctly.
