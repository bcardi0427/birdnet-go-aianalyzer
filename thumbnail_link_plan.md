# Plan: Configure Thumbnail Click Actions to External Bird Databases

This document outlines a plan to implement a configurable setting allowing users to decide what happens when clicking a bird thumbnail on the main dashboard and tables.

## 1. Feature Overview
Currently, clicking a bird thumbnail on the dashboard tables navigates the user to the internal detection details page. 
This feature will introduce a new setting: **"Click link destination"** (or similar) under the User Interface settings. Users will be able to choose from the following click actions for thumbnails:
1. **Details (Default)**: Navigates to the internal detection detail view.
2. **eBird**: Opens the species profile page on eBird in a new tab.
3. **Wikipedia**: Opens the species article on Wikipedia in a new tab.
4. **All About Birds**: Opens the species guide on All About Birds in a new tab.
5. **None**: Clicking does nothing (the image is static).

---

## 2. Configuration Settings Changes

### A. Go Backend Changes
We need to add a new `ClickLinkTo` field to the `Thumbnails` struct inside the configuration file.

1. **Modify `Thumbnails` struct** in [config.go](file:///F:/AntiGravity%20Sources/birdnet-go/internal/conf/config.go):
   ```go
   type Thumbnails struct {
       Debug          bool   `yaml:"debug" json:"debug"`
       Summary        bool   `yaml:"summary" json:"summary"`
       Recent         bool   `yaml:"recent" json:"recent"`
       ImageProvider  string `yaml:"imageprovider" json:"imageProvider"`
       FallbackPolicy string `yaml:"fallbackpolicy" json:"fallbackPolicy"`
       ClickLinkTo    string `yaml:"clicklinkto" json:"clickLinkTo"` // NEW field: "details", "ebird", "wikipedia", "allaboutbirds", "none"
   }
   ```

2. **Add default configuration** in [defaults.go](file:///F:/AntiGravity%20Sources/birdnet-go/internal/conf/defaults.go):
   ```go
   viper.SetDefault("realtime.dashboard.thumbnails.clicklinkto", "details")
   ```

---

### B. Frontend settings changes
We need to update the frontend type definitions so they sync correctly with the backend.

1. **Update `Thumbnails` interface** in [settings.ts](file:///F:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/stores/settings.ts):
   ```typescript
   export interface Thumbnails {
     debug?: boolean;
     summary: boolean;
     recent: boolean;
     imageProvider: string;
     fallbackPolicy: string;
     clickLinkTo: 'details' | 'ebird' | 'wikipedia' | 'allaboutbirds' | 'none'; // NEW field
   }
   ```

---

## 3. UI Settings Page Changes
We will expose the option in the **Visual Content** tab of the User Interface Settings page.

1. **Add dropdown option** in [UserInterfaceSettingsPage.svelte](file:///F:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/desktop/features/settings/pages/UserInterfaceSettingsPage.svelte) inside the **Bird Images** section:
   ```html
   <div class="grid grid-cols-1 md:grid-cols-2 gap-x-6">
     <SelectDropdown
       options={[
         { value: 'details', label: 'Detection Details' },
         { value: 'ebird', label: 'eBird' },
         { value: 'wikipedia', label: 'Wikipedia' },
         { value: 'allaboutbirds', label: 'All About Birds' },
         { value: 'none', label: 'None (Static Image)' }
       ]}
       value={settings.dashboard.thumbnails.clickLinkTo || 'details'}
       label="Thumbnail click destination"
       helpText="Choose what happens when clicking a bird thumbnail in tables and lists."
       disabled={store.isLoading || store.isSaving}
       variant="select"
       groupBy={false}
       menuSize="sm"
       onChange={value => updateThumbnailSetting('clickLinkTo', value as string)}
     />
   </div>
   ```

---

## 4. Frontend Link Helper
On the frontend, we can write a helper function to resolve the target URL in Svelte components. Since the frontend `Detection` and `DailySpeciesSummary` types already contain `scientificName` and `speciesCode` (the eBird code), we have all the information required on the client side:

```typescript
export function getBirdSiteLink(
  destination: string, 
  scientificName: string, 
  commonName: string, 
  speciesCode: string
): string | null {
  if (destination === 'wikipedia') {
    return `https://wikipedia.org/wiki/${encodeURIComponent(scientificName.replace(/ /g, '_'))}`;
  }
  if (destination === 'allaboutbirds') {
    const cleanedCommon = commonName.replace(/'/g, '');
    return `https://allaboutbirds.org/guide/${encodeURIComponent(cleanedCommon.replace(/ /g, '_'))}`;
  }
  if (destination === 'ebird' && speciesCode) {
    return `https://ebird.org/species/${encodeURIComponent(speciesCode)}`;
  }
  return null;
}
```

---

## 5. UI Rendering Changes
We need to update the thumbnail elements in tables and lists to act as links or buttons according to the configuration.

### A. Update `DetectionRow.svelte`
In [DetectionRow.svelte](file:///F:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/desktop/features/detections/components/DetectionRow.svelte):

1. **Retrieve setting and link**:
   ```typescript
   let clickLinkTo = $derived($settingsStore.formData.realtime?.dashboard?.thumbnails?.clickLinkTo ?? 'details');
   let externalLink = $derived(getBirdSiteLink(clickLinkTo, detection.scientificName, detection.commonName, detection.speciesCode));
   ```

2. **Render wrapper conditional link/button**:
   ```html
   {#if clickLinkTo === 'none'}
     <!-- Static thumbnail -->
     <div class="sp-thumbnail-button cursor-default">
       <img ... />
     </div>
   {:else}
     <!-- Dynamic click target -->
     <svelte:element
       this={externalLink ? 'a' : 'button'}
       href={externalLink}
       target={externalLink ? '_blank' : undefined}
       rel={externalLink ? 'noopener noreferrer' : undefined}
       onclick={externalLink ? undefined : handleDetailsClick}
       class="sp-thumbnail-button"
     >
       <img ... />
     </svelte:element>
   {/if}
   ```

### B. Update `DailySummaryCard.svelte` & `Search.svelte`
We will apply a similar template modification to:
- The daily summary table thumbnails in [DailySummaryCard.svelte](file:///F:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/desktop/features/dashboard/components/DailySummaryCard.svelte).
- The search result thumbnails in [Search.svelte](file:///F:/AntiGravity%20Sources/birdnet-go/frontend/src/lib/desktop/views/Search.svelte).
