<script lang="ts">
  import ChevronDown from '@lucide/svelte/icons/chevron-down';
  import { m } from '../lib/i18n';

  let { notes }: { notes: string[] } = $props();
</script>

<section class="card sheet space-y-3" aria-labelledby="h-notes">
  <h3 id="h-notes" class="card-title">{m.notes.title}</h3>
  <p class="max-w-prose text-sm text-ink-muted">{m.notes.intro}</p>
  <ol class="max-w-prose list-decimal space-y-2 pl-5 text-sm marker:text-ink-muted">
    {#each Object.values(m.notes.items) as item (item)}
      <li class="pl-1">{item}</li>
    {/each}
  </ol>
  <details class="disclosure text-caption">
    <summary>{m.notes.engine}<ChevronDown /></summary>
    <ul class="mt-2 list-disc space-y-1 pl-5 text-ink-muted">
      {#each notes as n, i (i)}<li>{n}</li>{/each}
    </ul>
  </details>
</section>
