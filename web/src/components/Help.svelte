<script lang="ts">
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import CircleHelp from '@lucide/svelte/icons/circle-help';
  import { m } from '../lib/i18n';
</script>

<details class="disclosure card no-print scroll-mt-24" id="help">
  <summary class="w-full text-ink"><CircleHelp class="text-link" /><span class="flex-1">{m.help.title}</span><ChevronDown /></summary>
  <ol class="mt-4 max-w-prose list-decimal space-y-2 pl-5 text-sm marker:font-semibold marker:text-link">
    {#each Object.values(m.help.steps) as step (step)}
      <li class="pl-1">{step}</li>
    {/each}
  </ol>
  <p class="mt-4 max-w-prose text-caption text-ink-muted">{m.print.howTo}</p>
</details>
