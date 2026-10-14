<script lang="ts">
  import type { Snippet } from 'svelte';

  let { kind = 'info', children }: { kind?: 'info' | 'warn' | 'error'; children: Snippet } = $props();
</script>

<div class="callout callout-{kind}" role={kind === 'error' ? 'alert' : undefined}>{@render children()}</div>
