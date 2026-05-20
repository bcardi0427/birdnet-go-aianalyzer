<script lang="ts">
  import { onMount, type Component } from 'svelte';
  import { api } from '$lib/utils/api';
  import { loggers } from '$lib/utils/logger';
  import {
    Clock,
    ExternalLink,
    Globe2,
    MapPin,
    RefreshCw,
    Route,
    ShieldCheck,
    Users,
  } from '@lucide/svelte';
  import type { IconProps } from '@lucide/svelte';

  const logger = loggers.settings;

  interface VisitorCount {
    key: string;
    count: number;
  }

  interface VisitorEntry {
    time: string;
    method: string;
    path: string;
    query: string;
    status: number;
    ip: string;
    host: string;
    referer: string;
    user_agent: string;
    cf_country: string;
    cf_ray: string;
    x_forwarded_proto: string;
    tunneled: boolean;
    authenticated: boolean;
    latency_ms: number;
  }

  interface VisitorResponse {
    entries: VisitorEntry[];
    totalReturned: number;
    uniqueIps: number;
    uniqueCountries: number;
    topIps: VisitorCount[];
    topCountries: VisitorCount[];
    topReferers: VisitorCount[];
    topPaths: VisitorCount[];
    logPath: string;
  }

  let loading = $state(true);
  let error = $state('');
  let visitors = $state<VisitorResponse>({
    entries: [],
    totalReturned: 0,
    uniqueIps: 0,
    uniqueCountries: 0,
    topIps: [],
    topCountries: [],
    topReferers: [],
    topPaths: [],
    logPath: 'logs/visitor.log',
  });

  const recentEntries = $derived([...visitors.entries].reverse());
  const loggedInVisits = $derived(visitors.entries.filter(entry => entry.authenticated).length);
  const guestVisits = $derived(visitors.entries.filter(entry => !entry.authenticated).length);

  onMount(() => {
    loadVisitors();
  });

  async function loadVisitors() {
    loading = true;
    error = '';
    try {
      visitors = await api.get<VisitorResponse>('/api/v2/system/visitors?limit=250');
    } catch (err) {
      logger.error('Failed to load visitor log', err);
      error = 'Could not load visitor log. Make sure you are logged in as an admin.';
    } finally {
      loading = false;
    }
  }

  function formatTime(value: string): string {
    if (!value) return 'Unknown';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleString();
  }

  function shortUserAgent(value: string): string {
    if (!value) return 'Unknown';
    if (value.length <= 90) return value;
    return `${value.slice(0, 87)}...`;
  }

  function displayReferer(value: string): string {
    if (!value) return 'Direct / blank';
    try {
      const url = new URL(value);
      return `${url.hostname}${url.pathname === '/' ? '' : url.pathname}`;
    } catch {
      return value;
    }
  }
</script>

<svelte:head>
  <title>Visitor Log</title>
</svelte:head>

<section class="space-y-6">
  <div
    class="rounded-xl border border-[var(--color-base-200)] bg-[var(--color-base-100)] shadow-sm overflow-hidden"
  >
    <div
      class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between px-6 py-5 border-b border-[var(--color-base-200)]"
    >
      <div class="flex items-start gap-4">
        <div
          class="rounded-xl bg-[var(--color-primary)]/10 text-[var(--color-primary)] p-3 shrink-0"
        >
          <Users class="size-6" />
        </div>
        <div>
          <h1 class="text-2xl font-bold text-[var(--color-base-content)]">Visitor Log</h1>
          <p class="text-sm text-[var(--color-base-content)]/65 mt-1">
            Admin-only view of public page visits, referrers, Cloudflare metadata, and recent
            visitor activity.
          </p>
          <p class="text-xs text-[var(--color-base-content)]/50 mt-2">
            Source: {visitors.logPath}
          </p>
        </div>
      </div>

      <button
        type="button"
        onclick={loadVisitors}
        disabled={loading}
        class="inline-flex items-center justify-center gap-2 rounded-lg border border-[var(--color-base-300)] px-4 py-2 text-sm font-medium text-[var(--color-base-content)] hover:bg-[var(--color-base-200)] disabled:opacity-50"
      >
        <RefreshCw class={`size-4 ${loading ? 'animate-spin' : ''}`} />
        Refresh
      </button>
    </div>

    {#if error}
      <div class="m-6 rounded-lg bg-[var(--color-error)]/10 text-[var(--color-error)] p-4">
        {error}
      </div>
    {:else}
      <div class="p-6 space-y-6">
        <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
          <div class="stat-card">
            <Users class="size-5 text-[var(--color-primary)]" />
            <div>
              <p class="stat-label">Recent page visits</p>
              <p class="stat-value">{visitors.totalReturned}</p>
            </div>
          </div>
          <div class="stat-card">
            <Globe2 class="size-5 text-[var(--color-info)]" />
            <div>
              <p class="stat-label">Unique IPs</p>
              <p class="stat-value">{visitors.uniqueIps}</p>
            </div>
          </div>
          <div class="stat-card">
            <MapPin class="size-5 text-[var(--color-success)]" />
            <div>
              <p class="stat-label">Countries</p>
              <p class="stat-value">{visitors.uniqueCountries}</p>
            </div>
          </div>
          <div class="stat-card">
            <ShieldCheck class="size-5 text-[var(--color-warning)]" />
            <div>
              <p class="stat-label">Guest / logged-in</p>
              <p class="stat-value">{guestVisits} / {loggedInVisits}</p>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 xl:grid-cols-2 gap-4">
          {@render SummaryList('Top pages', Route, visitors.topPaths)}
          {@render SummaryList('Top countries', MapPin, visitors.topCountries)}
          {@render SummaryList('Top IPs', Globe2, visitors.topIps)}
          {@render SummaryList('Top referrers', ExternalLink, visitors.topReferers)}
        </div>

        <div class="rounded-xl border border-[var(--color-base-200)] overflow-hidden">
          <div
            class="flex items-center gap-2 px-4 py-3 bg-[var(--color-base-200)]/45 border-b border-[var(--color-base-200)]"
          >
            <Clock class="size-4 text-[var(--color-base-content)]/70" />
            <h2 class="font-semibold text-[var(--color-base-content)]">Recent visits</h2>
          </div>

          {#if loading}
            <div class="p-6 text-sm text-[var(--color-base-content)]/60">Loading visitors...</div>
          {:else if recentEntries.length === 0}
            <div class="p-6 text-sm text-[var(--color-base-content)]/60">
              No visitor records found yet.
            </div>
          {:else}
            <div class="overflow-x-auto">
              <table class="min-w-full text-sm">
                <thead class="bg-[var(--color-base-200)]/35 text-left">
                  <tr>
                    <th class="table-heading">Time</th>
                    <th class="table-heading">IP / Country</th>
                    <th class="table-heading">Page</th>
                    <th class="table-heading">Referrer</th>
                    <th class="table-heading">User agent</th>
                    <th class="table-heading">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {#each recentEntries as entry}
                    <tr class="border-t border-[var(--color-base-200)] align-top">
                      <td class="table-cell whitespace-nowrap">{formatTime(entry.time)}</td>
                      <td class="table-cell">
                        <div class="font-medium">{entry.ip || 'Unknown'}</div>
                        <div class="text-xs text-[var(--color-base-content)]/55">
                          {entry.cf_country || 'Unknown'} · {entry.tunneled ? 'Cloudflare' : 'direct'}
                        </div>
                      </td>
                      <td class="table-cell">
                        <div class="font-medium">{entry.path}</div>
                        {#if entry.query}
                          <div class="text-xs text-[var(--color-base-content)]/55">?{entry.query}</div>
                        {/if}
                      </td>
                      <td class="table-cell max-w-xs break-words">{displayReferer(entry.referer)}</td>
                      <td class="table-cell max-w-md text-[var(--color-base-content)]/65">
                        {shortUserAgent(entry.user_agent)}
                      </td>
                      <td class="table-cell">
                        <span
                          class={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                            entry.status < 400
                              ? 'bg-[var(--color-success)]/10 text-[var(--color-success)]'
                              : 'bg-[var(--color-error)]/10 text-[var(--color-error)]'
                          }`}
                        >
                          {entry.status}
                        </span>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</section>

{#snippet SummaryList(title: string, icon: Component<IconProps>, items: VisitorCount[])}
  {@const SummaryIcon = icon}
  <div class="rounded-xl border border-[var(--color-base-200)] bg-[var(--color-base-100)] p-4">
    <div class="flex items-center gap-2 mb-3">
      <SummaryIcon class="size-4 text-[var(--color-primary)]" />
      <h2 class="font-semibold text-[var(--color-base-content)]">{title}</h2>
    </div>
    {#if items.length === 0}
      <p class="text-sm text-[var(--color-base-content)]/55">No data yet.</p>
    {:else}
      <div class="space-y-2">
        {#each items as item}
          <div class="flex items-start justify-between gap-3 text-sm">
            <span class="truncate text-[var(--color-base-content)]/75" title={item.key}>
              {item.key}
            </span>
            <span
              class="shrink-0 rounded-full bg-[var(--color-base-200)] px-2 py-0.5 text-xs font-medium"
            >
              {item.count}
            </span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/snippet}

<style>
  .stat-card {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    border: 1px solid var(--color-base-200);
    border-radius: 0.75rem;
    padding: 1rem;
    background: color-mix(in oklch, var(--color-base-100) 92%, var(--color-base-200));
  }

  .stat-label {
    font-size: 0.75rem;
    color: color-mix(in oklch, var(--color-base-content) 58%, transparent);
  }

  .stat-value {
    font-size: 1.5rem;
    font-weight: 800;
    line-height: 1.15;
    color: var(--color-base-content);
  }

  .table-heading {
    padding: 0.75rem 1rem;
    font-size: 0.75rem;
    font-weight: 700;
    color: color-mix(in oklch, var(--color-base-content) 70%, transparent);
  }

  .table-cell {
    padding: 0.85rem 1rem;
    color: var(--color-base-content);
  }
</style>
