<script lang="ts">
  import { m } from '../lib/i18n';
</script>

<details class="no-print" id="privacy">
  <summary class="cursor-pointer font-medium">{m.privacy.title}</summary>
  <ul class="mt-2 max-w-prose list-disc space-y-1 pl-5 text-sm">
    {#each Object.values(m.privacy.items) as item (item)}
      <li>{item}</li>
    {/each}
  </ul>
</details>
