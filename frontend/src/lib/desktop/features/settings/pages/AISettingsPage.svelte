<script lang="ts">
  import { onMount } from 'svelte';
  import { BrainCircuit, PlugZap, RefreshCw, Save } from '@lucide/svelte';
  import Checkbox from '$lib/desktop/components/forms/Checkbox.svelte';
  import NumberField from '$lib/desktop/components/forms/NumberField.svelte';
  import PasswordField from '$lib/desktop/components/forms/PasswordField.svelte';
  import SelectDropdown from '$lib/desktop/components/forms/SelectDropdown.svelte';
  import TextInput from '$lib/desktop/components/forms/TextInput.svelte';
  import LoadingSpinner from '$lib/desktop/components/ui/LoadingSpinner.svelte';
  import SettingsSection from '$lib/desktop/features/settings/components/SettingsSection.svelte';
  import type { SelectOption } from '$lib/desktop/components/forms/SelectDropdown.types';
  import { t } from '$lib/i18n';
  import { toastActions } from '$lib/stores/toast';
  import { settingsAPI, type AIModel, type AISettings } from '$lib/utils/settingsApi';

  const defaultSettings: AISettings = {
    enabled: false,
    apiKey: '',
    model: 'gemini-2.5-flash',
    reportDays: 1,
    cacheHours: 4,
    systemPrompt: '',
  };

  let settings = $state<AISettings>({ ...defaultSettings });
  let originalSettings = $state<AISettings>({ ...defaultSettings });
  let models = $state<AIModel[]>([]);
  let loading = $state(true);
  let saving = $state(false);
  let loadingModels = $state(false);
  let testingConnection = $state(false);
  let generatingFreshReport = $state(false);
  let error = $state<string | null>(null);
  let modelError = $state<string | null>(null);
  let connectionStatus = $state<{ ok: boolean; message: string } | null>(null);

  let hasChanges = $derived(JSON.stringify(settings) !== JSON.stringify(originalSettings));
  let modelOptions = $derived<SelectOption[]>([
    ...(settings.model && !models.some(model => model.id === settings.model)
      ? [{ value: settings.model, label: settings.model }]
      : []),
    ...models.map(model => ({
      value: model.id,
      label: model.displayName || model.id,
      description: model.description,
    })),
  ]);

  onMount(() => {
    loadSettings();
  });

  async function loadSettings() {
    loading = true;
    error = null;

    try {
      const data = await settingsAPI.ai.getSettings();
      settings = { ...defaultSettings, ...data };
      originalSettings = JSON.parse(JSON.stringify(settings));

      if (settings.apiKey && settings.apiKey !== '**********') {
        await loadModels();
      }
    } catch (err) {
      error = err instanceof Error ? err.message : t('settings.ai.errors.loadFailed');
    } finally {
      loading = false;
    }
  }

  async function loadModels() {
    if (!settings.apiKey) {
      modelError = t('settings.ai.errors.apiKeyRequiredForModels');
      return;
    }

    loadingModels = true;
    modelError = null;

    try {
      models = await settingsAPI.ai.getModels();
    } catch (err) {
      modelError = err instanceof Error ? err.message : t('settings.ai.errors.modelsFailed');
    } finally {
      loadingModels = false;
    }
  }

  async function saveSettings() {
    if (saving || !hasChanges) return;

    saving = true;
    error = null;

    try {
      const updated = await settingsAPI.ai.updateSettings(settings);
      settings = { ...defaultSettings, ...updated };
      originalSettings = JSON.parse(JSON.stringify(settings));
      toastActions.success(t('settings.ai.saved'));
    } catch (err) {
      error = err instanceof Error ? err.message : t('settings.ai.errors.saveFailed');
    } finally {
      saving = false;
    }
  }

  async function testConnection() {
    if (!settings.apiKey) {
      connectionStatus = { ok: false, message: t('settings.ai.errors.apiKeyRequiredForModels') };
      return;
    }

    testingConnection = true;
    connectionStatus = null;
    modelError = null;

    try {
      const availableModels = await settingsAPI.ai.getModels();
      const hasSelectedModel = availableModels.some(model => model.id === settings.model);
      connectionStatus = {
        ok: true,
        message: hasSelectedModel
          ? `Connection successful. Model "${settings.model}" is available.`
          : `Connection successful. Retrieved ${availableModels.length} model(s).`,
      };
      models = availableModels;
    } catch (err) {
      connectionStatus = {
        ok: false,
        message: err instanceof Error ? err.message : t('settings.ai.errors.modelsFailed'),
      };
    } finally {
      testingConnection = false;
    }
  }

  async function generateFreshReport() {
    if (hasChanges) {
      toastActions.warning('Save AI settings first before generating a fresh report.');
      return;
    }

    generatingFreshReport = true;
    try {
      const report = await settingsAPI.ai.getReportFresh();
      if (report.cached) {
        toastActions.info('Report request completed (cache was used).');
      } else {
        toastActions.success('Fresh AI report generated without cache.');
      }
    } catch (err) {
      toastActions.error(err instanceof Error ? err.message : 'Failed to generate fresh AI report.');
    } finally {
      generatingFreshReport = false;
    }
  }

  function resetChanges() {
    settings = JSON.parse(JSON.stringify(originalSettings));
  }
</script>

<main class="settings-page-content space-y-6" aria-label={t('settings.ai.title')}>
  <div>
    <div>
      <div class="flex items-center gap-3">
        <BrainCircuit class="size-7 text-[var(--color-primary)]" />
        <h1 class="text-2xl font-semibold text-[var(--color-base-content)]">
          {t('settings.ai.title')}
        </h1>
      </div>
      <p class="mt-2 max-w-2xl text-sm text-[var(--color-base-content)]/70">
        {t('settings.ai.description')}
      </p>
    </div>
  </div>

  {#if error}
    <div class="alert alert-error text-sm" role="alert">{error}</div>
  {/if}

  {#if loading}
    <div class="flex items-center justify-center py-16">
      <LoadingSpinner size="lg" />
    </div>
  {:else}
    <SettingsSection
      title={t('settings.ai.configuration.title')}
      description={t('settings.ai.configuration.description')}
      {hasChanges}
    >
      <div class="settings-form-grid">
        <Checkbox
          checked={settings.enabled}
          label={t('settings.ai.fields.enabled')}
          helpText={t('settings.ai.fields.enabledHelp')}
          onchange={value => (settings.enabled = value)}
        />

        <PasswordField
          label={t('settings.ai.fields.apiKey')}
          value={settings.apiKey}
          autocomplete="off"
          allowReveal={false}
          helpText={t('settings.ai.fields.apiKeyHelp')}
          onUpdate={value => (settings.apiKey = value)}
        />

        <div class="form-control">
          <label class="label" for="ai-test-connection">
            <span class="label-text">Test Gemini Connection</span>
          </label>
          <div class="flex items-center gap-2">
            <button
              id="ai-test-connection"
              type="button"
              class="btn btn-sm btn-outline gap-2"
              onclick={testConnection}
              disabled={testingConnection || !settings.apiKey}
            >
              {#if testingConnection}
                <LoadingSpinner size="xs" />
              {:else}
                <PlugZap class="size-4" />
              {/if}
              {testingConnection ? 'Testing…' : 'Test Connection'}
            </button>
            {#if connectionStatus}
              <span
                class={`text-sm ${connectionStatus.ok ? 'text-[var(--color-success)]' : 'text-[var(--color-error)]'}`}
                role="status"
              >
                {connectionStatus.message}
              </span>
            {/if}
          </div>
          <span class="help-text">Checks your API key and verifies Gemini is reachable.</span>
          <div class="mt-2">
            <button
              type="button"
              class="btn btn-sm btn-outline gap-2"
              onclick={generateFreshReport}
              disabled={generatingFreshReport || saving || hasChanges}
            >
              {#if generatingFreshReport}
                <LoadingSpinner size="xs" />
              {:else}
                <RefreshCw class="size-4" />
              {/if}
              {generatingFreshReport ? 'Generating…' : 'Generate Fresh Report (Bypass Cache)'}
            </button>
          </div>
          <span class="help-text">
            Generates a one-time report without reading or writing cache. Regular report views still use cache.
          </span>
        </div>

        <div class="form-control">
          <div class="flex items-end gap-2">
            <div class="min-w-0 flex-1">
              <SelectDropdown
                label={t('settings.ai.fields.model')}
                options={modelOptions}
                value={settings.model}
                searchable={true}
                placeholder={t('settings.ai.fields.modelPlaceholder')}
                disabled={loadingModels}
                onChange={value => {
                  if (typeof value === 'string') settings.model = value;
                }}
              />
            </div>
            <button
              type="button"
              class="btn btn-sm btn-outline mb-0.5"
              onclick={loadModels}
              disabled={loadingModels || !settings.apiKey}
            >
              {#if loadingModels}
                <LoadingSpinner size="sm" />
              {:else}
                <RefreshCw class="size-4" />
              {/if}
              {t('settings.ai.actions.refreshModels')}
            </button>
          </div>
          {#if modelError}
            <span class="help-text text-[var(--color-warning)]">{modelError}</span>
          {/if}
        </div>

        <NumberField
          label="Report days"
          value={settings.reportDays}
          min={1}
          max={31}
          step={1}
          helpText="How many days of detections to include in the AI report window (1-31)."
          onUpdate={value => (settings.reportDays = value)}
        />

        <NumberField
          label={t('settings.ai.fields.cacheHours')}
          value={settings.cacheHours}
          min={1}
          max={168}
          step={1}
          helpText={t('settings.ai.fields.cacheHoursHelp')}
          onUpdate={value => (settings.cacheHours = value)}
        />

        <TextInput
          label={t('settings.ai.fields.fallbackModel')}
          value={settings.model}
          placeholder="gemini-2.5-flash"
          helpText={t('settings.ai.fields.fallbackModelHelp')}
          oninput={value => (settings.model = value)}
        />
      </div>

      <div class="mt-6">
        <label class="label" for="ai-system-prompt">
          <span class="label-text">{t('settings.ai.fields.systemPrompt')}</span>
        </label>
        <textarea
          id="ai-system-prompt"
          class="textarea textarea-bordered min-h-32 w-full"
          bind:value={settings.systemPrompt}
          placeholder={t('settings.ai.fields.systemPromptPlaceholder')}
          maxlength="4000"
        ></textarea>
        <span class="help-text">{t('settings.ai.fields.systemPromptHelp')}</span>
      </div>
    </SettingsSection>

    <div
      class="mt-6 border-t border-[var(--color-base-300)] pt-4"
      role="toolbar"
      aria-label={t('settings.actions.toolbar')}
    >
      <div class="flex items-center justify-end gap-3">
        {#if hasChanges}
          <button
            type="button"
            class="btn btn-ghost btn-sm gap-2"
            onclick={resetChanges}
            disabled={saving}
            aria-label={t('settings.actions.resetAriaLabel')}
          >
            <RefreshCw class="size-4" aria-hidden="true" />
            {t('settings.actions.reset')}
          </button>
        {/if}
        <button
          type="button"
          class="btn btn-primary btn-sm gap-2"
          onclick={saveSettings}
          disabled={!hasChanges || saving}
          aria-busy={saving}
          aria-label={saving
            ? t('settings.actions.savingAriaLabel')
            : t('settings.actions.saveAriaLabel')}
        >
          {#if saving}
            <LoadingSpinner size="xs" />
            {t('settings.actions.saving')}
          {:else}
            <Save class="size-4" aria-hidden="true" />
            {t('settings.actions.save')}
          {/if}
        </button>
      </div>
    </div>
  {/if}
</main>
