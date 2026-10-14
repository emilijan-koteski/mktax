<script lang="ts">
  import { fmtAmount, fmtDate } from '../lib/format';
  import { m } from '../lib/i18n';
  import type { ImportSummary } from '../lib/types';

  let { summary }: { summary: ImportSummary } = $props();
  const ignored = $derived(Object.entries(summary.ignored).sort());
  const unknown = $derived(Object.entries(summary.unknown).sort());
</script>

<section class="card sheet space-y-6" aria-labelledby="h-excluded">
  <h3 id="h-excluded" class="card-title">{m.excluded.title}</h3>

  <div class="space-y-2">
    <h4 class="text-sm font-semibold">{m.excluded.interest}</h4>
    <p class="max-w-prose text-caption text-ink-muted">{m.excluded.interestHint}</p>
    {#if summary.interest.length}
      <div class="table-wrap max-w-xl">
        <table class="ledger ledger-stack">
          <thead><tr><th scope="col">{m.portal.date}</th><th scope="col">{m.excluded.action}</th><th scope="col" class="num">{m.excluded.amount}</th></tr></thead>
          <tbody>
            {#each summary.interest as r, i (i)}
              <tr>
                <td class="cell-head tabular-nums">{fmtDate(r.date)}</td>
                <td data-label={m.excluded.action}>{r.action}</td>
                <td data-label={m.excluded.amount} class="num">{fmtAmount(r.amount, r.currency)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <p class="text-sm">{m.common.none}</p>
    {/if}
  </div>

  {#if summary.excluded.length}
    <div class="space-y-2">
      <h4 class="text-sm font-semibold">{m.excluded.excludedDividends}</h4>
      <p class="max-w-prose text-caption text-ink-muted">{m.excluded.excludedHint}</p>
      <div class="table-wrap max-w-2xl">
        <table class="ledger ledger-stack">
          <thead><tr><th scope="col">{m.portal.date}</th><th scope="col">{m.portal.instrument}</th><th scope="col">{m.excluded.action}</th><th scope="col" class="num">{m.excluded.amount}</th></tr></thead>
          <tbody>
            {#each summary.excluded as r, i (i)}
              <tr>
                <td class="cell-head tabular-nums">{fmtDate(r.date)}</td>
                <td class="cell-head">{r.ticker}</td>
                <td data-label={m.excluded.action}>{r.action}</td>
                <td data-label={m.excluded.amount} class="num">{fmtAmount(r.amount, r.currency)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}

  {#if ignored.length}
    <div class="space-y-2">
      <h4 class="text-sm font-semibold">{m.excluded.ignored}</h4>
      <p class="max-w-prose text-caption text-ink-muted">{m.excluded.ignoredHint}</p>
      <ul class="flex flex-wrap gap-2">
        {#each ignored as [action, n] (action)}<li class="chip">{action}: {n}</li>{/each}
      </ul>
    </div>
  {/if}

  {#if unknown.length}
    <div class="space-y-2">
      <h4 class="text-sm font-semibold">{m.excluded.unknown}</h4>
      <p class="max-w-prose text-caption text-ink-muted">{m.excluded.unknownHint}</p>
      <ul class="flex flex-wrap gap-2">
        {#each unknown as [action, n] (action)}<li class="chip">{action}: {n}</li>{/each}
      </ul>
    </div>
  {/if}
</section>
