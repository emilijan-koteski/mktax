<script lang="ts">
  import { fmtDate, fmtMkd } from '../lib/format';
  import { m } from '../lib/i18n';
  import type { PortalRow } from '../lib/types';
  import CopyValue from './CopyValue.svelte';

  let { rows }: { rows: PortalRow[] } = $props();
</script>

<h4 class="mb-2 text-sm font-semibold">{m.month.enterInPortal}</h4>
<div class="table-wrap">
  <table class="ledger ledger-stack">
    <thead>
      <tr>
        <th scope="col">{m.portal.n}</th>
        <th scope="col">{m.portal.instrument}</th>
        <th scope="col">{m.portal.incomeType}</th>
        <th scope="col">{m.portal.incomeSubtype}</th>
        <th scope="col">{m.portal.country}</th>
        <th scope="col">{m.portal.currency}</th>
        <th scope="col" class="num">{m.portal.gross}</th>
        <th scope="col" class="num">{m.portal.foreignTax}</th>
        <th scope="col">{m.portal.date}</th>
        <th scope="col" class="num">{m.csv.headers.taxMkd}</th>
      </tr>
    </thead>
    <tbody>
      {#each rows as r (r.n)}
        <tr class:is-muted={!r.enter}>
          <th scope="row" class="cell-head tabular-nums text-ink-muted">{r.n}</th>
          <td class="cell-head font-medium">
            {r.ticker}
            <span class="sub">{r.kind === 'sale' ? m.portal.sale : m.portal.dividend}</span>
          </td>
          <td data-label={m.portal.incomeType}>{r.income_type}</td>
          <td data-label={m.portal.incomeSubtype}>{r.income_subtype}</td>
          <td data-label={m.portal.country}>{r.country.name_mk}</td>
          <td data-label={m.portal.currency}>{r.currency}</td>
          <td data-label={m.portal.gross} class="num">
            {#if r.enter}
              <CopyValue value={r.gross} />
            {:else}
              <span>{r.gross}</span>
              <span class="sub">{r.skip_reason ? m.portal.skip[r.skip_reason] : ''}</span>
            {/if}
          </td>
          <td data-label={m.portal.foreignTax} class="num">
            {#if r.enter && r.kind === 'dividend'}
              <CopyValue value={r.foreign_tax} />
            {:else}
              {r.foreign_tax}
            {/if}
          </td>
          <td data-label={m.portal.date} class="tabular-nums">{fmtDate(r.date)}</td>
          <td data-label={m.csv.headers.taxMkd} class="num">{fmtMkd(r.tax_mkd)}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
