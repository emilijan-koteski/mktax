<script lang="ts">
  import { fmtDate, fmtMkd } from '../lib/format';
  import { m } from '../lib/i18n';
  import type { PortalRow } from '../lib/types';
  import CopyValue from './CopyValue.svelte';

  let { rows }: { rows: PortalRow[] } = $props();
</script>

<div class="overflow-x-auto">
  <table class="ledger">
    <caption class="pb-1 text-left text-sm font-medium">{m.month.enterInPortal}</caption>
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
        <tr class={r.enter ? '' : 'text-muted'}>
          <th scope="row">{r.n}</th>
          <td>
            {r.ticker}
            <span class="block text-xs text-muted">{r.kind === 'sale' ? m.portal.sale : m.portal.dividend}</span>
          </td>
          <td>{r.income_type}</td>
          <td>{r.income_subtype}</td>
          <td>{r.country.name_mk}</td>
          <td>{r.currency}</td>
          <td class="num">
            {#if r.enter}
              <CopyValue value={r.gross} />
            {:else}
              <span>{r.gross}</span>
              <span class="block text-xs">{r.skip_reason ? m.portal.skip[r.skip_reason] : ''}</span>
            {/if}
          </td>
          <td class="num">
            {#if r.enter && r.kind === 'dividend'}
              <CopyValue value={r.foreign_tax} />
            {:else}
              {r.foreign_tax}
            {/if}
          </td>
          <td>{fmtDate(r.date)}</td>
          <td class="num">{fmtMkd(r.tax_mkd)}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
