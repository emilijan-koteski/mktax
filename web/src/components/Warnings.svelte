<script lang="ts">
  import { m } from '../lib/i18n';
  import type { Warning } from '../lib/types';
  import { warningText } from '../lib/warnings';
  import Callout from './Callout.svelte';

  let { warnings = [], raw = [] }: { warnings?: Warning[]; raw?: string[] } = $props();
</script>

{#if warnings.length || raw.length}
  <Callout kind="warn">
    <ul class="list-disc space-y-1 pl-5">
      {#each warnings as w, i (i)}
        <li>{warningText(w)}</li>
      {/each}
    </ul>
    {#if raw.length}
      <details class="mt-1 text-xs text-muted">
        <summary class="cursor-pointer">{m.common.technical}</summary>
        <ul class="list-disc pl-5">
          {#each raw as line, i (i)}<li>{line}</li>{/each}
        </ul>
      </details>
    {/if}
  </Callout>
{/if}
