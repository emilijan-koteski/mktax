<script lang="ts">
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import { fmtMonth } from '../lib/format';
  import { m } from '../lib/i18n';

  let { months, period = $bindable() }: { months: string[]; period: { from: string; to: string } } = $props();
</script>

<fieldset class="min-w-0">
  <legend class="mb-2 text-sm font-semibold">{m.review.period}</legend>
  <div class="grid grid-cols-2 gap-3 sm:flex">
    <label class="field">
      {m.review.from}
      <span class="select">
        <select class="sm:min-w-44" bind:value={period.from}>
          {#each months as mo (mo)}<option value={mo}>{fmtMonth(mo)}</option>{/each}
        </select>
        <ChevronDown />
      </span>
    </label>
    <label class="field">
      {m.review.to}
      <span class="select">
        <select class="sm:min-w-44" bind:value={period.to}>
          {#each months as mo (mo)}<option value={mo}>{fmtMonth(mo)}</option>{/each}
        </select>
        <ChevronDown />
      </span>
    </label>
  </div>
</fieldset>
