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
  import { appState } from '$lib/stores/appState.svelte';
  import {
    settingsAPI,
    type AIModel,
    type AISettings,
    type AIProviderSettings,
  } from '$lib/utils/settingsApi';

  const defaultProviderSettings = {
    apiKey: '',
    baseUrl: '',
    model: '',
  };

  const defaultSettings: AISettings = {
    enabled: false,
    provider: 'gemini',
    apiKey: '',
    baseUrl: '',
    model: 'gemini-2.5-flash',
    reportDays: 1,
    cacheHours: 4,
    systemPrompt: '',
    utmParameters: '',
    gemini: { ...defaultProviderSettings, model: 'gemini-2.5-flash' },
    openai: {
      ...defaultProviderSettings,
      model: 'gpt-4o-mini',
      baseUrl: 'https://api.openai.com/v1',
    },
    openrouter: {
      ...defaultProviderSettings,
      model: 'openai/gpt-4o-mini',
      baseUrl: 'https://openrouter.ai/api/v1',
    },
    openaiCompatible: { ...defaultProviderSettings },
    ollama: { ...defaultProviderSettings, model: 'llama3.2', baseUrl: 'http://localhost:11434/v1' },
    anthropic: { ...defaultProviderSettings, model: 'claude-3-5-haiku-latest' },
  };

  function getProviderKey(
    provider: string
  ): 'gemini' | 'openai' | 'openrouter' | 'openaiCompatible' | 'ollama' | 'anthropic' {
    if (provider === 'openai-compatible') return 'openaiCompatible';
    return provider as any;
  }

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
  const providerOptions: SelectOption[] = [
    { value: 'gemini', label: 'Google Gemini' },
    { value: 'openai', label: 'OpenAI' },
    { value: 'openrouter', label: 'OpenRouter' },
    { value: 'openai-compatible', label: 'OpenAI-compatible' },
    { value: 'ollama', label: 'Ollama' },
    { value: 'anthropic', label: 'Anthropic' },
  ];

  const providerDefaults: Record<string, { model: string; baseUrl: string }> = {
    gemini: { model: 'gemini-2.5-flash', baseUrl: '' },
    openai: { model: 'gpt-4o-mini', baseUrl: 'https://api.openai.com/v1' },
    openrouter: { model: 'openai/gpt-4o-mini', baseUrl: 'https://openrouter.ai/api/v1' },
    'openai-compatible': { model: '', baseUrl: '' },
    ollama: { model: 'llama3.2', baseUrl: 'http://localhost:11434/v1' },
    anthropic: { model: 'claude-3-5-haiku-latest', baseUrl: '' },
  };

  let showBaseUrl = $derived(
    settings.provider === 'openai-compatible' || settings.provider === 'ollama'
  );
  let requiresApiKey = $derived(
    settings.provider !== 'ollama' && settings.provider !== 'openai-compatible'
  );
  let activeModel = $derived(settings[getProviderKey(settings.provider)]?.model || '');
  let modelOptions = $derived<SelectOption[]>([
    ...(activeModel && !models.some(model => model.id === activeModel)
      ? [{ value: activeModel, label: activeModel }]
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
      settings = {
        ...defaultSettings,
        ...data,
        gemini: { ...defaultSettings.gemini, ...data.gemini } as AIProviderSettings,
        openai: { ...defaultSettings.openai, ...data.openai } as AIProviderSettings,
        openrouter: { ...defaultSettings.openrouter, ...data.openrouter } as AIProviderSettings,
        openaiCompatible: {
          ...defaultSettings.openaiCompatible,
          ...data.openaiCompatible,
        } as AIProviderSettings,
        ollama: { ...defaultSettings.ollama, ...data.ollama } as AIProviderSettings,
        anthropic: { ...defaultSettings.anthropic, ...data.anthropic } as AIProviderSettings,
      };

      // Ensure defaults are populated for all providers if they are empty
      for (const provider of [
        'gemini',
        'openai',
        'openrouter',
        'openai-compatible',
        'ollama',
        'anthropic',
      ]) {
        const key = getProviderKey(provider);
        const pSettings = settings[key];
        if (pSettings) {
          const defaults = providerDefaults[provider] ?? { model: '', baseUrl: '' };
          if (!pSettings.model) {
            pSettings.model = defaults.model;
          }
          if (!pSettings.baseUrl) {
            pSettings.baseUrl = defaults.baseUrl;
          }
        }
      }

      originalSettings = JSON.parse(JSON.stringify(settings));

      const activeKey = settings[getProviderKey(settings.provider)]?.apiKey;
      if (!requiresApiKey || activeKey) {
        await loadModels();
      }
    } catch (err) {
      error = err instanceof Error ? err.message : t('settings.ai.errors.loadFailed');
    } finally {
      loading = false;
    }
  }

  async function loadModels() {
    if (hasChanges) {
      modelError = 'Save AI settings first to load models for the selected provider.';
      return;
    }

    const activeKey = settings[getProviderKey(settings.provider)]?.apiKey;
    if (requiresApiKey && !activeKey) {
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
      const activeKey = getProviderKey(settings.provider);
      const active = settings[activeKey];
      if (active) {
        if (!active.baseUrl?.trim()) {
          if (settings.provider === 'openai') active.baseUrl = providerDefaults.openai.baseUrl;
          if (settings.provider === 'openrouter')
            active.baseUrl = providerDefaults.openrouter.baseUrl;
          if (settings.provider === 'ollama') active.baseUrl = providerDefaults.ollama.baseUrl;
        }

        const normalizedModel = String(active.model || '')
          .trim()
          .toLowerCase();
        const modelLooksGemini =
          normalizedModel.startsWith('gemini') || normalizedModel.startsWith('models/gemini');
        if (!active.model || (settings.provider !== 'gemini' && modelLooksGemini)) {
          active.model = providerDefaults[settings.provider]?.model || active.model;
        }

        settings.apiKey = active.apiKey;
        settings.baseUrl = active.baseUrl;
        settings.model = active.model;
      }

      // Send a plain JSON object (not a reactive proxy) to ensure all fields,
      // especially provider/baseUrl, are persisted correctly.
      const payload: AISettings = JSON.parse(JSON.stringify(settings));

      // Guard against dropdown binding edge-cases: always persist a normalized provider.
      payload.provider = String(payload.provider || settings.provider || 'gemini')
        .trim()
        .toLowerCase();
      if (!payload.provider) payload.provider = 'gemini';

      const updated = await settingsAPI.ai.updateSettings(payload);
      settings = {
        ...defaultSettings,
        ...updated,
        gemini: { ...defaultSettings.gemini, ...updated.gemini } as AIProviderSettings,
        openai: { ...defaultSettings.openai, ...updated.openai } as AIProviderSettings,
        openrouter: { ...defaultSettings.openrouter, ...updated.openrouter } as AIProviderSettings,
        openaiCompatible: {
          ...defaultSettings.openaiCompatible,
          ...updated.openaiCompatible,
        } as AIProviderSettings,
        ollama: { ...defaultSettings.ollama, ...updated.ollama } as AIProviderSettings,
        anthropic: { ...defaultSettings.anthropic, ...updated.anthropic } as AIProviderSettings,
      };
      originalSettings = JSON.parse(JSON.stringify(settings));
      toastActions.success(t('settings.ai.saved'));
    } catch (err) {
      error = err instanceof Error ? err.message : t('settings.ai.errors.saveFailed');
    } finally {
      saving = false;
    }
  }

  async function testConnection() {
    if (hasChanges) {
      connectionStatus = {
        ok: false,
        message: 'Save AI settings first to test the selected provider configuration.',
      };
      return;
    }

    const activeKey = settings[getProviderKey(settings.provider)]?.apiKey;
    if (requiresApiKey && !activeKey) {
      connectionStatus = { ok: false, message: t('settings.ai.errors.apiKeyRequiredForModels') };
      return;
    }

    testingConnection = true;
    connectionStatus = null;
    modelError = null;

    try {
      const availableModels = await settingsAPI.ai.getModels();
      const activeModel = settings[getProviderKey(settings.provider)]?.model;
      const hasSelectedModel = availableModels.some(model => model.id === activeModel);
      connectionStatus = {
        ok: true,
        message: hasSelectedModel
          ? `Connection successful. Model "${activeModel}" is available.`
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
      toastActions.error(
        err instanceof Error ? err.message : 'Failed to generate fresh AI report.'
      );
    } finally {
      generatingFreshReport = false;
    }
  }

  function resetChanges() {
    settings = JSON.parse(JSON.stringify(originalSettings));
  }

  function onProviderChange(next: string) {
    settings.provider = next;
    // Clear connection status and models
    connectionStatus = null;
    modelError = null;
    models = [];

    const nextKey = getProviderKey(next);
    const nextSettings = settings[nextKey];
    if (nextSettings) {
      const nextDefaults = providerDefaults[next] ?? { model: '', baseUrl: '' };
      if (!nextSettings.model) {
        nextSettings.model = nextDefaults.model;
      }
      if (!nextSettings.baseUrl) {
        nextSettings.baseUrl = nextDefaults.baseUrl;
      }
    }
  }

  function handleProviderSelect(value: string | string[]) {
    if (typeof value !== 'string') return;
    if (value === settings.provider) return;
    onProviderChange(value);
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
      <div
        class="mb-3 rounded border border-[var(--color-base-300)] bg-[var(--color-base-200)]/40 px-3 py-2 text-xs text-[var(--color-base-content)]/80"
      >
        Runtime build: <span class="font-mono">{appState.version}</span>
      </div>

      <div class="settings-form-grid">
        <Checkbox
          checked={settings.enabled}
          label={t('settings.ai.fields.enabled')}
          helpText={t('settings.ai.fields.enabledHelp')}
          onchange={value => (settings.enabled = value)}
        />

        <SelectDropdown
          label="Provider"
          options={providerOptions}
          bind:value={settings.provider}
          placeholder="Select provider"
          onChange={handleProviderSelect}
        />

        {#key settings.provider}
          <PasswordField
            label={t('settings.ai.fields.apiKey')}
            bind:value={settings[getProviderKey(settings.provider)]!.apiKey}
            autocomplete="off"
            allowReveal={false}
            helpText={settings.provider === 'openrouter'
              ? 'OpenRouter API key. Required unless using a keyless local gateway.'
              : settings.provider === 'openai'
                ? 'OpenAI API key.'
                : settings.provider === 'anthropic'
                  ? 'Anthropic API key.'
                  : settings.provider === 'ollama'
                    ? 'Optional for local Ollama unless your gateway requires it.'
                    : settings.provider === 'openai-compatible'
                      ? 'API key for your compatible gateway, if required.'
                      : 'Google AI Studio API key.'}
          />

          <div class="form-control">
            <label class="label" for="ai-test-connection">
              <span class="label-text">Test AI Provider Connection</span>
            </label>
            <div class="flex items-center gap-2">
              <button
                id="ai-test-connection"
                type="button"
                class="btn btn-sm btn-outline gap-2"
                onclick={testConnection}
                disabled={testingConnection ||
                  (requiresApiKey && !settings[getProviderKey(settings.provider)]?.apiKey)}
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
            <span class="help-text"
              >Checks provider access and verifies the configured endpoint is reachable.</span
            >
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
              Generates a one-time report without reading or writing cache. Regular report views
              still use cache.
            </span>
          </div>

          <div class="form-control">
            <div class="flex items-end gap-2">
              <div class="min-w-0 flex-1">
                <SelectDropdown
                  label={t('settings.ai.fields.model')}
                  options={modelOptions}
                  bind:value={settings[getProviderKey(settings.provider)]!.model}
                  searchable={true}
                  placeholder={t('settings.ai.fields.modelPlaceholder')}
                  disabled={loadingModels}
                />
              </div>
              <button
                type="button"
                class="btn btn-sm btn-outline mb-0.5"
                onclick={loadModels}
                disabled={loadingModels ||
                  hasChanges ||
                  (requiresApiKey && !settings[getProviderKey(settings.provider)]?.apiKey)}
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

          {#if showBaseUrl}
            <TextInput
              label="Base URL"
              bind:value={settings[getProviderKey(settings.provider)]!.baseUrl}
              placeholder={settings.provider === 'ollama'
                ? 'http://localhost:11434/v1'
                : 'https://your-provider/v1'}
              helpText={settings.provider === 'openai-compatible'
                ? 'Required for OpenAI-compatible providers.'
                : 'Optional override for Ollama endpoint.'}
            />
          {/if}
        {/key}

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
          label="UTM Parameters"
          bind:value={settings.utmParameters}
          placeholder="utm_source=birdnet-go&utm_medium=report"
          helpText="Optional UTM/tracking parameters to append to generated report links (e.g. bcardi0427 or utm_source=birdnet-go)."
        />

        {#key settings.provider}
          <TextInput
            label={t('settings.ai.fields.fallbackModel')}
            bind:value={settings[getProviderKey(settings.provider)]!.model}
            placeholder={providerDefaults[settings.provider]?.model || 'Enter model ID'}
            helpText={t('settings.ai.fields.fallbackModelHelp')}
          />
        {/key}
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
