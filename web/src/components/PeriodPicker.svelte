<script lang="ts">
  import { fmtMonth } from '../lib/format';
  import { m } from '../lib/i18n';

  let { months, period = $bindable() }: { months: string[]; period: { from: string; to: string } } = $props();
</script>

<div class="flex flex-wrap items-center gap-3 text-sm">
  <span class="font-medium">{m.review.period}</span>
  <label class="inline-flex items-center gap-1">
    {m.review.from}
    <select class="rounded-xs border border-line bg-sheet px-2 py-1" bind:value={period.from}>
      {#each months as mo (mo)}<option value={mo}>{fmtMonth(mo)}</option>{/each}
    </select>
  </label>
  <label class="inline-flex items-center gap-1">
    {m.review.to}
    <select class="rounded-xs border border-line bg-sheet px-2 py-1" bind:value={period.to}>
      {#each months as mo (mo)}<option value={mo}>{fmtMonth(mo)}</option>{/each}
    </select>
  </label>
</div>
