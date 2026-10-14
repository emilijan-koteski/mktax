<script lang="ts">
  import { fmtAmount, fmtDate, fmtMkd, fmtQty } from '../lib/format';
  import { m, t } from '../lib/i18n';
  import type { Sale } from '../lib/types';
  import Callout from './Callout.svelte';

  let { sale, index }: { sale: Sale; index: number } = $props();
</script>

<section class="space-y-2">
  <h4 class="text-sm font-semibold">
    {index}. {t(m.lots.saleTotal, { quantity: fmtQty(sale.quantity), ticker: sale.ticker, proceeds: fmtAmount(sale.proceeds, sale.currency), rate: sale.sale_rate })}
  </h4>
  {#if !sale.cost_complete}
    <Callout kind="warn">{m.lots.costIncomplete}</Callout>
  {/if}
  <p class="text-xs font-medium text-ink-muted">{m.lots.title}</p>
  <div class="table-wrap">
    <table class="ledger ledger-stack">
      <thead>
        <tr>
          <th scope="col">{m.lots.acquired}</th>
          <th scope="col" class="num">{m.lots.quantity}</th>
          <th scope="col" class="num">{m.lots.cost}</th>
          <th scope="col" class="num">{m.lots.proceeds}</th>
          <th scope="col" class="num">{m.lots.acquisitionRate}</th>
          <th scope="col" class="num">{m.lots.gainA}</th>
          <th scope="col" class="num">{m.lots.gainB}</th>
          <th scope="col">{m.lots.status}</th>
          <th scope="col">{m.lots.taxFreeFrom}</th>
        </tr>
      </thead>
      <tbody>
        {#each sale.lots as lot, i (i)}
          <tr class:is-muted={lot.status !== 'taxable'}>
            <td class="cell-head tabular-nums">{fmtDate(lot.acquired)}</td>
            <td data-label={m.lots.quantity} class="num">{fmtQty(lot.quantity)}</td>
            <td data-label={m.lots.cost} class="num">{lot.status === 'unknown_cost' ? '?' : fmtAmount(lot.cost, lot.currency)}</td>
            <td data-label={m.lots.proceeds} class="num">{fmtAmount(lot.proceeds, sale.currency)}</td>
            <td data-label={m.lots.acquisitionRate} class="num">{lot.acquisition_rate}</td>
            <td data-label={m.lots.gainA} class="num">{lot.status === 'unknown_cost' ? '' : fmtMkd(lot.gain_mkd)}</td>
            <td data-label={m.lots.gainB} class="num">{lot.status === 'unknown_cost' ? '' : fmtMkd(lot.gain_mkd_b)}</td>
            <td data-label={m.lots.status}>{m.lots.statuses[lot.status]}</td>
            <td data-label={m.lots.taxFreeFrom} class="tabular-nums">{fmtDate(lot.tax_free_from)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  <p class="text-xs text-ink-muted">
    {m.lots.saleRate}: {sale.sale_rate};
    {m.totals.taxableGains}: {fmtMkd(sale.taxable_gain_mkd)};
    {m.lots.gainB}: {fmtMkd(sale.taxable_gain_mkd_b)};
    {m.totals.capitalGainsTax}: {fmtMkd(sale.tax_mkd)}
  </p>
</section>
