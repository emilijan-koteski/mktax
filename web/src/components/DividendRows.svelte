<script lang="ts">
  import { fmtAmount, fmtDate, fmtMkd } from '../lib/format';
  import { m } from '../lib/i18n';
  import type { Dividend } from '../lib/types';

  let { dividends }: { dividends: Dividend[] } = $props();
  const estimated = (d: Dividend) => d.warnings.some((w) => w.code === 'gross_estimated');
  const anyEstimated = $derived(dividends.some(estimated));
</script>

<section class="space-y-2">
  <h4 class="text-sm font-semibold">{m.dividends.title}</h4>
  <div class="table-wrap">
    <table class="ledger ledger-stack">
      <thead>
        <tr>
          <th scope="col">{m.portal.instrument}</th>
          <th scope="col">{m.portal.date}</th>
          <th scope="col" class="num">{m.dividends.gross}</th>
          <th scope="col" class="num">{m.dividends.net}</th>
          <th scope="col" class="num">{m.dividends.withheld}</th>
          <th scope="col" class="num">{m.dividends.rate}</th>
          <th scope="col" class="num">{m.dividends.grossMkd}</th>
          <th scope="col" class="num">{m.dividends.taxAccepted}</th>
          <th scope="col" class="num">{m.dividends.taxRejected}</th>
        </tr>
      </thead>
      <tbody>
        {#each dividends as d, i (i)}
          <tr>
            <th scope="row" class="cell-head">{d.ticker} <span class="sub">{d.type}</span></th>
            <td data-label={m.portal.date} class="tabular-nums">{fmtDate(d.date)}</td>
            <td data-label={m.dividends.gross} class="num">
              {#if estimated(d)}<span class="text-accent-ink" title={m.dividends.estimated}>≈</span>{/if}
              {fmtAmount(d.gross, d.gross_currency)}
            </td>
            <td data-label={m.dividends.net} class="num">{fmtAmount(d.net, d.net_currency)}</td>
            <td data-label={m.dividends.withheld} class="num">{fmtAmount(d.foreign_tax, d.foreign_tax_currency)}</td>
            <td data-label={m.dividends.rate} class="num">{d.gross_rate}</td>
            <td data-label={m.dividends.grossMkd} class="num">{fmtMkd(d.gross_mkd)}</td>
            <td data-label={m.dividends.taxAccepted} class="num">{fmtMkd(d.tax_if_credit_accepted_mkd)}</td>
            <td data-label={m.dividends.taxRejected} class="num">{fmtMkd(d.tax_if_credit_rejected_mkd)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  {#if anyEstimated}<p class="text-xs text-ink-muted">≈ {m.dividends.estimated}</p>{/if}
</section>
