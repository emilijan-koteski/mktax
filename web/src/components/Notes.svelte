<script lang="ts">
  import { m } from '../lib/i18n';

  let { notes }: { notes: string[] } = $props();
</script>

<section class="sheet mt-6" aria-labelledby="h-notes">
  <h3 id="h-notes" class="text-xl">{m.notes.title}</h3>
  <p class="mt-1 max-w-prose text-sm text-muted">{m.notes.intro}</p>
  <ol class="mt-2 max-w-prose list-decimal space-y-1 pl-5 text-sm">
    {#each Object.values(m.notes.items) as item (item)}
      <li>{item}</li>
    {/each}
  </ol>
  <details class="mt-3 text-xs text-muted">
    <summary class="cursor-pointer">{m.notes.engine}</summary>
    <ul class="mt-1 list-disc space-y-1 pl-5">
      {#each notes as n, i (i)}<li>{n}</li>{/each}
    </ul>
  </details>
</section>
