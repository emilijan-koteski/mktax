<script lang="ts">
  import { m } from '../lib/i18n';
  import type { Instrument } from '../lib/types';

  let { instruments, fundUnits = $bindable() }: { instruments: Instrument[]; fundUnits: Record<string, boolean> } = $props();
</script>

<div class="overflow-x-auto">
  <table class="ledger">
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
          <th scope="row">{ins.ticker}</th>
          <td>{ins.name}</td>
          <td class="text-muted">{ins.isin}</td>
          <td class="num">{ins.buys} / {ins.sells}</td>
          <td>
            <label class="inline-flex items-center gap-2">
              <input type="checkbox" aria-label="{m.instruments.fundUnits}: {ins.ticker}" bind:checked={fundUnits[ins.key]} />
              {#if ins.fund_units_guessed}<span class="text-xs text-muted">{m.instruments.guessed}</span>{/if}
            </label>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
