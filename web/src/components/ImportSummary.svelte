<script lang="ts">
  import FileText from '@lucide/svelte/icons/file-text';
  import X from '@lucide/svelte/icons/x';
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

<ul class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
  {#each imp.files as f (f.name)}
    <li class="flex min-w-0 items-center gap-3 rounded-md border border-line bg-surface py-2.5 pr-2 pl-3">
      <span class="grid size-9 shrink-0 place-items-center rounded-sm bg-brand-soft text-link"><FileText class="size-[18px]" /></span>
      <span class="min-w-0 flex-1">
        <span class="block truncate font-semibold" title={f.name}>{f.name}</span>
        <span class="flex flex-wrap gap-x-2 text-xs text-ink-muted">
          <span>{t(m.files.rows, { n: f.rows })}</span>
          <span>{t(m.files.period, { from: fmtDate(f.first), to: fmtDate(f.last) })}</span>
        </span>
      </span>
      <button type="button" class="btn btn-ghost btn-icon shrink-0 text-ink-muted" aria-label="{m.files.remove} {f.name}" title={m.files.remove} onclick={() => onremove(f.name)}>
        <X />
      </button>
    </li>
  {/each}
</ul>

<ul class="mt-4 flex flex-wrap gap-2">
  <li class="chip">{imp.counts.trades} {m.review.trades}</li>
  <li class="chip">{imp.counts.dividends} {m.review.dividends}</li>
  <li class="chip">{imp.interest.length} {m.review.interest}</li>
  <li class="chip">{m.review.ignored}: {ignoredTotal}</li>
  <li class="chip">{m.review.accountCurrency}: {imp.account_currency}</li>
</ul>

{#if imp.flags.length}
  <div class="mt-4">
    <Callout kind="warn">
      <ul class="list-disc space-y-1 pl-4">
        {#each grouped as line (line)}
          <li>{line}</li>
        {/each}
      </ul>
    </Callout>
  </div>
{/if}
