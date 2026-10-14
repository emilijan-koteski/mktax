<script lang="ts">
  import { fmtAmount, fmtDate, fmtMkd, fmtQty } from '../lib/format';
  import { m, t } from '../lib/i18n';
  import type { Sale } from '../lib/types';
  import Callout from './Callout.svelte';

  let { sale, index }: { sale: Sale; index: number } = $props();
</script>

<section class="mt-3">
  <h4 class="text-sm font-semibold">
    {index}. {t(m.lots.saleTotal, { quantity: fmtQty(sale.quantity), ticker: sale.ticker, proceeds: fmtAmount(sale.proceeds, sale.currency), rate: sale.sale_rate })}
  </h4>
  {#if !sale.cost_complete}
    <div class="mt-1"><Callout kind="warn">{m.lots.costIncomplete}</Callout></div>
  {/if}
  <p class="mb-1 text-xs text-muted">{m.lots.title}</p>
  <div class="overflow-x-auto">
    <table class="ledger">
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
          <tr class={lot.status === 'taxable' ? '' : 'text-muted'}>
            <td>{fmtDate(lot.acquired)}</td>
            <td class="num">{fmtQty(lot.quantity)}</td>
            <td class="num">{lot.status === 'unknown_cost' ? '?' : fmtAmount(lot.cost, lot.currency)}</td>
            <td class="num">{fmtAmount(lot.proceeds, sale.currency)}</td>
            <td class="num">{lot.acquisition_rate}</td>
            <td class="num">{lot.status === 'unknown_cost' ? '' : fmtMkd(lot.gain_mkd)}</td>
            <td class="num">{lot.status === 'unknown_cost' ? '' : fmtMkd(lot.gain_mkd_b)}</td>
            <td>{m.lots.statuses[lot.status]}</td>
            <td>{fmtDate(lot.tax_free_from)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  <p class="mt-1 text-xs text-muted">
    {m.lots.saleRate}: {sale.sale_rate};
    {m.totals.taxableGains}: {fmtMkd(sale.taxable_gain_mkd)};
    {m.lots.gainB}: {fmtMkd(sale.taxable_gain_mkd_b)};
    {m.totals.capitalGainsTax}: {fmtMkd(sale.tax_mkd)}
  </p>
</section>
