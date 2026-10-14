<script lang="ts">
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import TriangleAlert from '@lucide/svelte/icons/triangle-alert';
  import { m } from '../lib/i18n';
</script>

<aside class="callout callout-warn no-print">
  <TriangleAlert />
  <div class="callout-body">
    <strong class="font-semibold">{m.disclaimer.short}</strong>
    <details class="disclosure mt-1">
      <summary class="text-caption">{m.disclaimer.more}<ChevronDown /></summary>
      <p class="mt-2 max-w-prose">{m.disclaimer.full}</p>
    </details>
  </div>
</aside>
