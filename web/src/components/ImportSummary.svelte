<script lang="ts">
  import { fmtDate } from '../lib/format';
  import { m, t } from '../lib/i18n';
  import type { ParseResult } from '../lib/types';
  import { warningText } from '../lib/warnings';
  import Callout from './Callout.svelte';

  let { info, onremove }: { info: ParseResult; onremove: (name: string) => void } = $props();
  const imp = $derived(info.import);
  const ignoredTotal = $derived(Object.values(imp.ignored).reduce((a, b) => a + b, 0));
  const grouped = $derived.by(() => {
    const byCode: Record<string, { text: string; n: number }> = {};
    for (const w of imp.flags) {
      if (byCode[w.code]) byCode[w.code].n += 1;
      else byCode[w.code] = { text: warningText(w), n: 1 };
    }
    return Object.values(byCode).map((g) => (g.n > 1 ? t(m.warnings.grouped, { text: g.text, n: g.n }) : g.text));
  });
</script>

<ul class="divide-y divide-line border-y border-line">
  {#each imp.files as f (f.name)}
    <li class="flex flex-wrap items-center gap-x-4 gap-y-1 py-2 text-sm">
      <span class="font-medium">{f.name}</span>
      <span class="text-muted">{t(m.files.rows, { n: f.rows })}</span>
      <span class="text-muted">{t(m.files.period, { from: fmtDate(f.first), to: fmtDate(f.last) })}</span>
      <button type="button" class="btn btn-secondary ml-auto" onclick={() => onremove(f.name)}>{m.files.remove}</button>
    </li>
  {/each}
</ul>

<p class="mt-3 text-sm">
  <span>{imp.counts.trades} {m.review.trades}</span>,
  <span>{imp.counts.dividends} {m.review.dividends}</span>,
  <span>{imp.interest.length} {m.review.interest}</span>;
  <span>{m.review.ignored}: {ignoredTotal}</span>;
  <span>{m.review.accountCurrency}: {imp.account_currency}</span>
</p>

{#if imp.flags.length}
  <div class="mt-3">
    <Callout kind="warn">
      <ul class="list-disc space-y-1 pl-5">
        {#each grouped as line (line)}
          <li>{line}</li>
        {/each}
      </ul>
    </Callout>
  </div>
{/if}
