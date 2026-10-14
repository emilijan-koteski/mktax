<script lang="ts">
  import { fmtAmount, fmtDate, fmtQty } from '../lib/format';
  import { m, t } from '../lib/i18n';
  import type { OpenPosition } from '../lib/types';

  let { positions, asOf }: { positions: OpenPosition[]; asOf: string } = $props();
</script>

<section class="sheet mt-6" aria-labelledby="h-open-lots">
  <h3 id="h-open-lots" class="text-xl">{m.openLots.title}</h3>
  <p class="text-sm text-muted">{t(m.openLots.asOf, { date: fmtDate(asOf) })}</p>
  {#if positions.length === 0}
    <p class="mt-2 text-sm">{m.openLots.none}</p>
  {:else}
    <div class="mt-2 overflow-x-auto">
      <table class="ledger">
        <thead>
          <tr>
            <th scope="col">{m.portal.instrument}</th>
            <th scope="col" class="num">{m.openLots.quantity}</th>
            <th scope="col" class="num">{m.openLots.cost}</th>
            <th scope="col">{m.openLots.earliest}</th>
            <th scope="col">{m.openLots.taxFreeFrom}</th>
            <th scope="col" class="num">{m.openLots.taxFreeNow}</th>
          </tr>
        </thead>
        <tbody>
          {#each positions as p (p.key)}
            <tr>
              <th scope="row">{p.ticker} <span class="block text-xs font-normal text-muted">{p.name}</span></th>
              <td class="num">{fmtQty(p.quantity)}</td>
              <td class="num">{fmtAmount(p.cost, p.currency)}</td>
              <td>{fmtDate(p.earliest_acquired)}</td>
              <td>{fmtDate(p.oldest_lot_tax_free_from)}</td>
              <td class="num">{fmtQty(p.tax_free_quantity_now)}</td>
            </tr>
            {#each p.lots as lot, i (i)}
              <tr class="text-xs text-muted">
                <td class="pl-4">{fmtDate(lot.acquired)}</td>
                <td class="num">{fmtQty(lot.quantity)}</td>
                <td class="num">{fmtAmount(lot.cost, p.currency)}</td>
                <td></td>
                <td>{fmtDate(lot.tax_free_from)}</td>
                <td></td>
              </tr>
            {/each}
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</section>
