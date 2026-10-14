<script lang="ts">
  import { fmtAmount, fmtDate, fmtMkd } from '../lib/format';
  import { m } from '../lib/i18n';
  import type { Dividend } from '../lib/types';

  let { dividends }: { dividends: Dividend[] } = $props();
  const estimated = (d: Dividend) => d.warnings.some((w) => w.code === 'gross_estimated');
  const anyEstimated = $derived(dividends.some(estimated));
</script>

<section class="mt-3">
  <p class="mb-1 text-xs text-muted">{m.dividends.title}</p>
  <div class="overflow-x-auto">
    <table class="ledger">
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
            <th scope="row">{d.ticker} <span class="block text-xs font-normal text-muted">{d.type}</span></th>
            <td>{fmtDate(d.date)}</td>
            <td class="num">
              {#if estimated(d)}<span title={m.dividends.estimated}>≈</span>{/if}
              {fmtAmount(d.gross, d.gross_currency)}
            </td>
            <td class="num">{fmtAmount(d.net, d.net_currency)}</td>
            <td class="num">{fmtAmount(d.foreign_tax, d.foreign_tax_currency)}</td>
            <td class="num">{d.gross_rate}</td>
            <td class="num">{fmtMkd(d.gross_mkd)}</td>
            <td class="num">{fmtMkd(d.tax_if_credit_accepted_mkd)}</td>
            <td class="num">{fmtMkd(d.tax_if_credit_rejected_mkd)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
  {#if anyEstimated}<p class="mt-1 text-xs text-muted">≈ {m.dividends.estimated}</p>{/if}
</section>
