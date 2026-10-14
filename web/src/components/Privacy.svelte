<script lang="ts">
  import Check from '@lucide/svelte/icons/check';
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import ShieldCheck from '@lucide/svelte/icons/shield-check';
  import { m } from '../lib/i18n';
</script>

<details class="disclosure card no-print scroll-mt-24" id="privacy">
  <summary class="w-full text-ink"><ShieldCheck class="text-link" /><span class="flex-1">{m.privacy.title}</span><ChevronDown /></summary>
  <ul class="mt-4 max-w-prose space-y-2 text-sm">
    {#each Object.values(m.privacy.items) as item (item)}
      <li class="flex gap-2"><Check class="mt-0.5 size-4 shrink-0 text-success" /><span>{item}</span></li>
    {/each}
  </ul>
</details>
