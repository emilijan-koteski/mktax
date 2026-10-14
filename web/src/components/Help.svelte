<script lang="ts">
  import { m } from '../lib/i18n';
</script>

<details class="no-print" id="help">
  <summary class="cursor-pointer font-medium">{m.help.title}</summary>
  <ol class="mt-2 max-w-prose list-decimal space-y-1 pl-5 text-sm">
    {#each Object.values(m.help.steps) as step (step)}
      <li>{step}</li>
    {/each}
  </ol>
  <p class="mt-2 max-w-prose text-sm text-muted">{m.print.howTo}</p>
</details>
