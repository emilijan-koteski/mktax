<script lang="ts">
  import { fmtAmount, fmtDate, fmtQty } from '../lib/format';
  import { m, t } from '../lib/i18n';
  import type { OpenPosition } from '../lib/types';

  let { positions, asOf }: { positions: OpenPosition[]; asOf: string } = $props();
</script>

<section class="card sheet space-y-4" aria-labelledby="h-open-lots">
  <header>
    <h3 id="h-open-lots" class="card-title">{m.openLots.title}</h3>
    <p class="text-caption text-ink-muted">{t(m.openLots.asOf, { date: fmtDate(asOf) })}</p>
  </header>
  {#if positions.length === 0}
    <p class="text-sm">{m.openLots.none}</p>
  {:else}
    <div class="table-wrap">
      <table class="ledger ledger-stack">
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
              <th scope="row" class="cell-head">{p.ticker} <span class="sub">{p.name}</span></th>
              <td data-label={m.openLots.quantity} class="num">{fmtQty(p.quantity)}</td>
              <td data-label={m.openLots.cost} class="num">{fmtAmount(p.cost, p.currency)}</td>
              <td data-label={m.openLots.earliest} class="tabular-nums">{fmtDate(p.earliest_acquired)}</td>
              <td data-label={m.openLots.taxFreeFrom} class="tabular-nums">{fmtDate(p.oldest_lot_tax_free_from)}</td>
              <td data-label={m.openLots.taxFreeNow} class="num">{fmtQty(p.tax_free_quantity_now)}</td>
            </tr>
            {#each p.lots as lot, i (i)}
              <tr class="is-child">
                <td data-label={m.lots.acquired} class="tabular-nums">{fmtDate(lot.acquired)}</td>
                <td data-label={m.openLots.quantity} class="num">{fmtQty(lot.quantity)}</td>
                <td data-label={m.openLots.cost} class="num">{fmtAmount(lot.cost, p.currency)}</td>
                <td></td>
                <td data-label={m.openLots.taxFreeFrom} class="tabular-nums">{fmtDate(lot.tax_free_from)}</td>
                <td></td>
              </tr>
            {/each}
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</section>
