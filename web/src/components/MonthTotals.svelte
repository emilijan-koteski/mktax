<script lang="ts">
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import { fmtMkd, isZero } from '../lib/format';
  import { m } from '../lib/i18n';
  import type { Totals } from '../lib/types';

  let { totals }: { totals: Totals } = $props();
  const rows = $derived(
    [
      [m.totals.taxableGains, totals.taxable_gains_mkd],
      [m.totals.taxableLosses, totals.taxable_losses_mkd],
      [m.totals.taxBase, totals.tax_base_mkd],
      [m.totals.capitalGainsTax, totals.capital_gains_tax_mkd],
      [m.totals.exemptGains, totals.exempt_gains_mkd],
      [m.totals.outOfScopeGains, totals.out_of_scope_gains_mkd],
      [m.totals.dividendGross, totals.dividend_gross_mkd],
      [m.totals.dividendWithheld, totals.dividend_withheld_mkd],
      [m.totals.dividendTaxAccepted, totals.dividend_tax_if_credit_accepted_mkd],
      [m.totals.dividendTaxRejected, totals.dividend_tax_if_credit_rejected_mkd],
    ].filter(([, v]) => !isZero(v)),
  );
</script>

<div class="grid gap-x-8 gap-y-4 text-sm md:grid-cols-2">
  <dl class="grid grid-cols-[1fr_auto] gap-y-1.5">
    {#each rows as [label, value] (label)}
      <dt class="text-ink-muted">{label}</dt>
      <dd class="num pl-4">{fmtMkd(value)}</dd>
    {/each}
    <dt class="mt-1 border-t border-line pt-2.5 font-semibold">{m.totals.taxDueAccepted}</dt>
    <dd class="num mt-1 border-t border-line pt-2.5 pl-4 font-semibold">{fmtMkd(totals.tax_due_if_credit_accepted_mkd)}</dd>
    <dt class="font-semibold">{m.totals.taxDueRejected}</dt>
    <dd class="num pl-4 font-semibold">{fmtMkd(totals.tax_due_if_credit_rejected_mkd)}</dd>
  </dl>
  <details class="disclosure self-start rounded-md bg-surface-muted px-4 py-3">
    <summary>{m.totals.comparisons}<ChevronDown /></summary>
    <dl class="mt-3 grid grid-cols-[1fr_auto] gap-x-4 gap-y-1.5">
      <dt class="text-ink-muted">{m.totals.methodB}</dt>
      <dd class="num">{fmtMkd(totals.capital_gains_tax_mkd_b)}</dd>
      <dt class="text-ink-muted">{m.totals.lossesOffset}</dt>
      <dd class="num">{fmtMkd(totals.tax_if_losses_offset_mkd)}</dd>
    </dl>
    <div class="mt-3 space-y-2 text-xs text-ink-muted">
      <p>{m.totals.methodBHint}</p>
      <p>{m.totals.lossesOffsetHint}</p>
      <p>{m.totals.creditHint}</p>
    </div>
  </details>
</div>
