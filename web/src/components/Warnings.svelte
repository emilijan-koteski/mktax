<script lang="ts">
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import { m } from '../lib/i18n';
  import type { Warning } from '../lib/types';
  import { warningText } from '../lib/warnings';
  import Callout from './Callout.svelte';

  let { warnings = [], raw = [] }: { warnings?: Warning[]; raw?: string[] } = $props();
</script>

{#if warnings.length || raw.length}
  <Callout kind="warn">
    {#if warnings.length}
      <ul class="list-disc space-y-1 pl-4">
        {#each warnings as w, i (i)}
          <li>{warningText(w)}</li>
        {/each}
      </ul>
    {/if}
    {#if raw.length}
      <details class="disclosure mt-2 text-xs">
        <summary>{m.common.technical}<ChevronDown /></summary>
        <ul class="mt-1 list-disc pl-4 text-ink-muted">
          {#each raw as line, i (i)}<li>{line}</li>{/each}
        </ul>
      </details>
    {/if}
  </Callout>
{/if}
