<script lang="ts">
  import { onMount } from 'svelte';
import DOMPurify, { type Config as DOMPurifyConfig } from 'dompurify';
  import { marked } from 'marked';
  import { BrainCircuit, RefreshCw, Settings, Sparkles } from '@lucide/svelte';
  import LoadingSpinner from '$lib/desktop/components/ui/LoadingSpinner.svelte';
  import { t } from '$lib/i18n';
  import { navigation } from '$lib/stores/navigation.svelte';
  import { settingsAPI } from '$lib/utils/settingsApi';

  let report = $state('');
  let loading = $state(true);
  let loadingFresh = $state(false);
  let error = $state<string | null>(null);
  let refreshedAt = $state<Date | null>(null);
  let reportIsCached = $state(false);
  let reportDays = $state(1);

  const sanitizeConfig: DOMPurifyConfig = {
    ALLOWED_TAGS: [
      'p',
      'br',
      'strong',
      'em',
      'ul',
      'ol',
      'li',
      'blockquote',
      'code',
      'pre',
      'h1',
      'h2',
      'h3',
      'h4',
      'h5',
      'h6',
      'table',
      'thead',
      'tbody',
      'tr',
      'th',
      'td',
      'caption',
      'a',
      'img',
    ],
    ALLOWED_ATTR: ['href', 'title', 'rel', 'target', 'src', 'alt', 'loading', 'width', 'height'],
  };

  function sanitizeReportHtml(input: string): string {
    const sanitized = DOMPurify.sanitize(input, sanitizeConfig);

    const parser = new DOMParser();
    const doc = parser.parseFromString(sanitized, 'text/html');
    const images = Array.from(doc.querySelectorAll('img'));
    images.forEach(img => {
      const src = img.getAttribute('src') ?? '';
      if (!src.startsWith('/api/v2/media/')) img.remove();
    });

    promoteInlineNarrativeHeadings(doc);

    return doc.body.innerHTML;
  }

  function normalizeNarrativeHeadings(input: string): string {
    // Preferred generic marker for prompt-defined sections:
    // [[H3]] Your Section Title
    // This enables user-defined headings without code changes.
    let out = input.replace(
      /(^|\n)\s*\[\[H3\]\]\s*(.+?)\s*$/gim,
      (_m, sep, title) => `${sep}### ${String(title).trim()}`
    );

    // Backward-compatible fallback for legacy reports that still emit
    // known plain heading labels without markdown markers.
    const fallbackHeadings = [
      'Weekly Acoustic Overview',
      'Hour-by-Hour Activity Pattern',
      'Species-by-Habitat Grouping',
      'Closing Note',
    ];

    for (const heading of fallbackHeadings) {
      const escaped = heading.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      const atParagraphStart = new RegExp(`(^|\\n\\s*\\n)(?:#{1,6}\\s*)?(${escaped})(?=\\s+)`, 'gmi');
      out = out.replace(atParagraphStart, (_m, sep, label) => `${sep}### ${label}`);

      const splitInline = new RegExp(`(^|\\n\\s*\\n)(###\\s+${escaped})\\s+`, 'gmi');
      out = out.replace(splitInline, (_m, sep, label) => `${sep}${label}\n\n`);
    }
    return out;
  }

  function promoteInlineNarrativeHeadings(doc: Document) {
    const headingPrefixes = [
      'Weekly Acoustic Overview',
      'Hour-by-Hour Activity Pattern',
      'Species-by-Habitat Grouping',
      'Closing Note',
    ];

    const paragraphs = Array.from(doc.querySelectorAll('p'));
    for (const p of paragraphs) {
      const text = (p.textContent ?? '').trim();
      if (!text) continue;

      const matched = headingPrefixes.find(
        prefix => text === prefix || text.startsWith(prefix + ' ')
      );
      if (!matched) continue;

      const remainder = text.slice(matched.length).trim();
      const heading = doc.createElement('h3');
      heading.textContent = matched;
      p.before(heading);

      if (remainder.length > 0) {
        p.textContent = remainder;
      } else {
        p.remove();
      }
    }
  }

  let renderedReport = $derived(
    report
      ? sanitizeReportHtml(marked.parse(normalizeNarrativeHeadings(report), { async: false }) as string)
      : ''
  );

  onMount(() => {
    loadReport(false);
  });

  async function loadReportDays() {
    try {
      const aiSettings = await settingsAPI.ai.getSettings();
      const days = Number(aiSettings?.reportDays ?? 1);
      reportDays = Number.isFinite(days) && days > 0 ? Math.min(Math.floor(days), 31) : 1;
    } catch {
      reportDays = 1;
    }
  }

  async function loadReport(bypassCache = false) {
    if (bypassCache) loadingFresh = true;
    else loading = true;
    error = null;

    try {
      await loadReportDays();
      const response = bypassCache
        ? await settingsAPI.ai.getReportFresh()
        : await settingsAPI.ai.getReport();
      report = response.report;
      refreshedAt = response.generatedAt ? new Date(response.generatedAt) : new Date();
      reportIsCached = response.cached === true;
    } catch (err) {
      const message = err instanceof Error ? err.message.toLowerCase() : '';
      if (message.includes('disabled'))
        error = 'AI analysis is disabled. Enable it in Settings → AI.';
      else if (message.includes('api key'))
        error = 'Gemini API key is missing. Configure it in Settings → AI.';
      else if (message.includes('timeout'))
        error = 'AI report generation timed out. Please try again.';
      else
        error = err instanceof Error ? err.message : t('aiAnalysis.errors.loadFailed');
    } finally {
      if (bypassCache) loadingFresh = false;
      else loading = false;
    }
  }
</script>

<svelte:head>
  <title>{t('aiAnalysis.title')} - BirdNET-Go</title>
</svelte:head>

<main class="col-span-12 w-full space-y-6">
  <section
    class="overflow-hidden rounded-lg border border-[var(--border-100)] bg-[var(--color-base-100)]"
  >
    <div
      class="flex flex-col gap-5 border-b border-[var(--border-100)] p-5 md:flex-row md:items-center md:justify-between"
    >
      <div class="flex min-w-0 items-start gap-4">
        <div
          class="flex size-12 shrink-0 items-center justify-center rounded-lg bg-[var(--color-primary)]/10 text-[var(--color-primary)]"
        >
          <BrainCircuit class="size-7" />
        </div>
        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="text-2xl font-semibold text-[var(--color-base-content)]">
              {t('aiAnalysis.title')}
            </h1>
            <span
              class="inline-flex items-center gap-1 rounded-md bg-[var(--color-accent)]/10 px-2 py-1 text-xs font-medium text-[var(--color-accent)]"
            >
              <Sparkles class="size-3.5" />
              {t('aiAnalysis.badge')}
            </span>
          </div>
          <p class="mt-2 max-w-3xl text-sm text-[var(--color-base-content)]/70">
            A generated summary of the last {reportDays} {reportDays === 1 ? 'day' : 'days'} of
            detected bird activity, formatted for quick review.
          </p>
          {#if refreshedAt}
            <p class="mt-2 text-xs text-[var(--color-base-content)]/50">
              {t('aiAnalysis.refreshedAt', { time: refreshedAt.toLocaleTimeString() })}
              {#if reportIsCached}
                <span class="ml-2 rounded bg-[var(--color-base-200)] px-1.5 py-0.5 text-[10px] uppercase tracking-wide">
                  cached
                </span>
              {/if}
            </p>
          {/if}
        </div>
      </div>

      <div class="flex shrink-0 items-center gap-2">
        <button
          type="button"
          class="btn btn-sm btn-outline"
          onclick={() => navigation.navigate('/ui/settings/ai')}
        >
          <Settings class="size-4" />
          {t('aiAnalysis.actions.settings')}
        </button>
        <button
          type="button"
          class="btn btn-sm btn-primary"
          onclick={() => loadReport(false)}
          disabled={loading || loadingFresh}
        >
          {#if loading}
            <LoadingSpinner size="sm" />
          {:else}
            <RefreshCw class="size-4" />
          {/if}
          {t('aiAnalysis.actions.refresh')}
        </button>
        <button
          type="button"
          class="btn btn-sm btn-outline"
          onclick={() => loadReport(true)}
          disabled={loading || loadingFresh}
        >
          {#if loadingFresh}
            <LoadingSpinner size="sm" />
          {:else}
            <RefreshCw class="size-4" />
          {/if}
          Refresh (Bypass Cache)
        </button>
      </div>
    </div>

    {#if loading}
      <div class="flex min-h-96 items-center justify-center">
        <div class="flex flex-col items-center gap-3">
          <LoadingSpinner size="lg" />
          <p class="text-sm text-[var(--color-base-content)]/60">{t('aiAnalysis.loading')}</p>
        </div>
      </div>
    {:else if error}
      <div class="p-5">
        <div class="alert alert-error text-sm" role="alert">{error}</div>
      </div>
    {:else if renderedReport}
      <article class="ai-report p-5 md:p-8">
        {@html renderedReport}
      </article>
    {:else}
      <div class="p-10 text-center text-[var(--color-base-content)]/60">
        {t('aiAnalysis.empty')}
      </div>
    {/if}
  </section>
</main>

<style>
  :global(.ai-report) {
    color: var(--color-base-content);
    line-height: 1.7;
  }

  :global(.ai-report > * + *) {
    margin-top: 1rem;
  }

  :global(.ai-report h1),
  :global(.ai-report h2),
  :global(.ai-report h3) {
    color: var(--color-base-content);
    font-weight: 700;
    line-height: 1.25;
  }

  :global(.ai-report h1) {
    font-size: 1.75rem;
  }

  :global(.ai-report h2) {
    margin-top: 1.75rem;
    font-size: 1.35rem;
  }

  :global(.ai-report h3) {
    margin-top: 1.35rem;
    font-size: 1.1rem;
  }

  :global(.ai-report ul),
  :global(.ai-report ol) {
    padding-left: 1.25rem;
  }

  :global(.ai-report li + li) {
    margin-top: 0.35rem;
  }

  :global(.ai-report blockquote) {
    border-left: 3px solid var(--color-primary);
    background: color-mix(in srgb, var(--color-primary) 8%, transparent);
    margin: 1.25rem 0;
    padding: 0.8rem 1rem;
    border-radius: 0.35rem;
  }

  :global(.ai-report table) {
    width: 100%;
    border-collapse: collapse;
    overflow: hidden;
    border-radius: 0.5rem;
  }

  :global(.ai-report th),
  :global(.ai-report td) {
    border: 1px solid var(--border-100);
    padding: 0.65rem 0.75rem;
    text-align: left;
  }

  :global(.ai-report th) {
    background: var(--color-base-200);
    font-weight: 600;
  }

  :global(.ai-report code) {
    border-radius: 0.25rem;
    background: var(--color-base-200);
    padding: 0.1rem 0.35rem;
    font-size: 0.9em;
  }
</style>
