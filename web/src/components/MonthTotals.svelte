<script lang="ts">
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

<div class="mt-3 grid gap-x-6 gap-y-1 text-sm sm:grid-cols-2">
  <dl class="grid grid-cols-[1fr_auto] gap-x-4 gap-y-0.5">
    {#each rows as [label, value] (label)}
      <dt class="text-muted">{label}</dt>
      <dd class="num">{fmtMkd(value)}</dd>
    {/each}
    <dt class="border-t border-ink pt-1 font-semibold">{m.totals.taxDueAccepted}</dt>
    <dd class="num border-t border-ink pt-1 font-semibold">{fmtMkd(totals.tax_due_if_credit_accepted_mkd)}</dd>
    <dt class="font-semibold">{m.totals.taxDueRejected}</dt>
    <dd class="num font-semibold">{fmtMkd(totals.tax_due_if_credit_rejected_mkd)}</dd>
  </dl>
  <details class="text-sm">
    <summary class="cursor-pointer text-accent">{m.totals.comparisons}</summary>
    <dl class="mt-1 grid grid-cols-[1fr_auto] gap-x-4 gap-y-0.5">
      <dt class="text-muted">{m.totals.methodB}</dt>
      <dd class="num">{fmtMkd(totals.capital_gains_tax_mkd_b)}</dd>
      <dt class="text-muted">{m.totals.lossesOffset}</dt>
      <dd class="num">{fmtMkd(totals.tax_if_losses_offset_mkd)}</dd>
    </dl>
    <p class="mt-1 max-w-prose text-xs text-muted">{m.totals.methodBHint}</p>
    <p class="mt-1 max-w-prose text-xs text-muted">{m.totals.lossesOffsetHint}</p>
    <p class="mt-1 max-w-prose text-xs text-muted">{m.totals.creditHint}</p>
  </details>
</div>
