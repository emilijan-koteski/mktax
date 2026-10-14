<script lang="ts">
  import { m } from '../lib/i18n';
  import type { Instrument } from '../lib/types';

  let { instruments, fundUnits = $bindable() }: { instruments: Instrument[]; fundUnits: Record<string, boolean> } = $props();
</script>

<div class="table-wrap">
  <table class="ledger ledger-stack">
    <thead>
      <tr>
        <th scope="col">{m.instruments.ticker}</th>
        <th scope="col">{m.instruments.name}</th>
        <th scope="col">{m.instruments.isin}</th>
        <th scope="col" class="num">{m.instruments.trades}</th>
        <th scope="col">{m.instruments.fundUnits}</th>
      </tr>
    </thead>
    <tbody>
      {#each instruments as ins (ins.key)}
        <tr>
          <th scope="row" class="cell-head">{ins.ticker}</th>
          <td data-label={m.instruments.name}>{ins.name}</td>
          <td data-label={m.instruments.isin} class="text-ink-muted tabular-nums">{ins.isin}</td>
          <td data-label={m.instruments.trades} class="num">{ins.buys} / {ins.sells}</td>
          <td data-label={m.instruments.fundUnits}>
            <label class="inline-flex items-center gap-2">
              <input type="checkbox" class="checkbox" aria-label="{m.instruments.fundUnits}: {ins.ticker}" bind:checked={fundUnits[ins.key]} />
              {#if ins.fund_units_guessed}<span class="text-xs text-ink-muted">{m.instruments.guessed}</span>{/if}
            </label>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
