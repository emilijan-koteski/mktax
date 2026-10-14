<script lang="ts">
  import Calendar from '@lucide/svelte/icons/calendar';
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
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

<section class="card sheet space-y-5" aria-labelledby="h-{month.month}">
  <header class="space-y-1.5">
    <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
      <h3 id="h-{month.month}" class="first-upper text-2xl">{fmtMonth(month.month)}</h3>
      {#if month.overdue}
        <span class="badge badge-danger">{m.month.status.overdue}</span>
      {:else if month.in_progress}
        <span class="badge badge-brand">{m.month.status.inProgress}</span>
      {:else}
        <span class="badge">{m.month.status.upcoming} {fmtDate(month.file_by)}</span>
      {/if}
    </div>
    <p class="flex items-center gap-2 text-caption text-ink-muted">
      <Calendar class="size-4 shrink-0" />{t(m.month.deadlines, { file: fmtDate(month.file_by), pay: fmtDate(month.pay_by) })}
    </p>
  </header>

  {#if month.has_entries}
    <div class="tax-due">
      <p class="tax-due-amount">{t(m.month.summary, { tax: fmtMkd(month.totals.tax_due_if_credit_accepted_mkd) })}</p>
      <p class="mt-0.5 text-caption text-ink-muted">
        {plural(month.totals.sales, m.month.sales)}, {plural(month.totals.dividends, m.month.dividends)}.
        {#if noCredit}{t(m.month.summaryNoCredit, { taxNoCredit: fmtMkd(month.totals.tax_due_if_credit_rejected_mkd) })}{/if}
      </p>
    </div>
    <Warnings warnings={month.warnings} raw={month.raw_warnings} />
    <div><PortalRows rows={month.portal_rows} /></div>
    <MonthTotals totals={month.totals} />
    <button type="button" class="btn btn-secondary btn-sm no-print" aria-expanded={showDetails} onclick={() => (open = !open)}>
      {showDetails ? m.month.hideDetails : m.month.showDetails}
      <ChevronDown class="transition-transform {showDetails ? 'rotate-180' : ''}" />
    </button>
    {#if showDetails}
      <div class="space-y-5 border-t border-line pt-5">
        {#each month.sales as sale, i (i)}<SaleDetails {sale} index={i + 1} />{/each}
        {#if month.dividends.length}<DividendRows dividends={month.dividends} />{/if}
      </div>
    {/if}
  {:else}
    <p class="text-sm text-ink-muted">{m.month.nothingToFile}</p>
  {/if}
</section>
