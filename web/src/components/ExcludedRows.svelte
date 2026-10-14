<script lang="ts">
  import { fmtAmount, fmtDate } from '../lib/format';
  import { m } from '../lib/i18n';
  import type { ImportSummary } from '../lib/types';

  let { summary }: { summary: ImportSummary } = $props();
  const ignored = $derived(Object.entries(summary.ignored).sort());
  const unknown = $derived(Object.entries(summary.unknown).sort());
</script>

<section class="sheet mt-6" aria-labelledby="h-excluded">
  <h3 id="h-excluded" class="text-xl">{m.excluded.title}</h3>

  <h4 class="mt-3 text-sm font-semibold">{m.excluded.interest}</h4>
  <p class="max-w-prose text-xs text-muted">{m.excluded.interestHint}</p>
  {#if summary.interest.length}
    <table class="ledger mt-1 max-w-md">
      <thead><tr><th scope="col">{m.portal.date}</th><th scope="col">{m.excluded.action}</th><th scope="col" class="num">{m.excluded.amount}</th></tr></thead>
      <tbody>
        {#each summary.interest as r, i (i)}
          <tr><td>{fmtDate(r.date)}</td><td>{r.action}</td><td class="num">{fmtAmount(r.amount, r.currency)}</td></tr>
        {/each}
      </tbody>
    </table>
  {:else}
    <p class="text-sm">{m.common.none}</p>
  {/if}

  {#if summary.excluded.length}
    <h4 class="mt-3 text-sm font-semibold">{m.excluded.excludedDividends}</h4>
    <p class="max-w-prose text-xs text-muted">{m.excluded.excludedHint}</p>
    <table class="ledger mt-1 max-w-xl">
      <thead><tr><th scope="col">{m.portal.date}</th><th scope="col">{m.portal.instrument}</th><th scope="col">{m.excluded.action}</th><th scope="col" class="num">{m.excluded.amount}</th></tr></thead>
      <tbody>
        {#each summary.excluded as r, i (i)}
          <tr><td>{fmtDate(r.date)}</td><td>{r.ticker}</td><td>{r.action}</td><td class="num">{fmtAmount(r.amount, r.currency)}</td></tr>
        {/each}
      </tbody>
    </table>
  {/if}

  {#if ignored.length}
    <h4 class="mt-3 text-sm font-semibold">{m.excluded.ignored}</h4>
    <p class="max-w-prose text-xs text-muted">{m.excluded.ignoredHint}</p>
    <ul class="mt-1 text-sm">
      {#each ignored as [action, n] (action)}<li>{action}: {n}</li>{/each}
    </ul>
  {/if}

  {#if unknown.length}
    <h4 class="mt-3 text-sm font-semibold">{m.excluded.unknown}</h4>
    <p class="max-w-prose text-xs text-muted">{m.excluded.unknownHint}</p>
    <ul class="mt-1 text-sm">
      {#each unknown as [action, n] (action)}<li>{action}: {n}</li>{/each}
    </ul>
  {/if}
</section>
