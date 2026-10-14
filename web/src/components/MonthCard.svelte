<script lang="ts">
  import { fmtDate, fmtMkd, fmtMonth } from '../lib/format';
  import { m, plural, t } from '../lib/i18n';
  import type { Month } from '../lib/types';
  import DividendRows from './DividendRows.svelte';
  import MonthTotals from './MonthTotals.svelte';
  import PortalRows from './PortalRows.svelte';
  import SaleDetails from './SaleDetails.svelte';
  import Warnings from './Warnings.svelte';

  let { month, printMode = false }: { month: Month; printMode?: boolean } = $props();
  let open = $state(false);
  const showDetails = $derived(open || printMode);
  const noCredit = $derived(month.totals.tax_due_if_credit_accepted_mkd !== month.totals.tax_due_if_credit_rejected_mkd);
</script>

<section class="sheet mt-6" aria-labelledby="h-{month.month}">
  <header class="flex flex-wrap items-baseline gap-x-4 gap-y-1">
    <h3 id="h-{month.month}" class="text-2xl">{fmtMonth(month.month)}</h3>
    {#if month.overdue}
      <span class="stamp">{m.month.status.overdue}</span>
    {:else if month.in_progress}
      <span class="badge badge-live">{m.month.status.inProgress}</span>
    {:else}
      <span class="badge">{m.month.status.upcoming} {fmtDate(month.file_by)}</span>
    {/if}
    <p class="basis-full text-sm text-muted">{t(m.month.deadlines, { file: fmtDate(month.file_by), pay: fmtDate(month.pay_by) })}</p>
  </header>

  {#if month.has_entries}
    <p class="mt-2 text-sm">
      {plural(month.totals.sales, m.month.sales)}, {plural(month.totals.dividends, m.month.dividends)}.
      <strong>{t(m.month.summary, { tax: fmtMkd(month.totals.tax_due_if_credit_accepted_mkd) })}</strong>
      {#if noCredit}<span class="text-muted">{t(m.month.summaryNoCredit, { taxNoCredit: fmtMkd(month.totals.tax_due_if_credit_rejected_mkd) })}</span>{/if}
    </p>
    <div class="mt-3"><Warnings warnings={month.warnings} raw={month.raw_warnings} /></div>
    <div class="mt-3"><PortalRows rows={month.portal_rows} /></div>
    <MonthTotals totals={month.totals} />
    <button type="button" class="btn btn-secondary no-print mt-3" aria-expanded={showDetails} onclick={() => (open = !open)}>
      {showDetails ? m.month.hideDetails : m.month.showDetails}
    </button>
    {#if showDetails}
      {#each month.sales as sale, i (i)}<SaleDetails {sale} index={i + 1} />{/each}
      {#if month.dividends.length}<DividendRows dividends={month.dividends} />{/if}
    {/if}
  {:else}
    <p class="mt-2 text-sm text-muted">{m.month.nothingToFile}</p>
  {/if}
</section>
