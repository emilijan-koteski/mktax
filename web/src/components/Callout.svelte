<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import Info from '@lucide/svelte/icons/info';
  import TriangleAlert from '@lucide/svelte/icons/triangle-alert';
  import type { Snippet } from 'svelte';

  let { kind = 'info', children }: { kind?: 'info' | 'warn' | 'error'; children: Snippet } = $props();
  const Icon = $derived(kind === 'error' ? CircleAlert : kind === 'warn' ? TriangleAlert : Info);
</script>

<div class="callout callout-{kind}" role={kind === 'error' ? 'alert' : undefined}>
  <Icon />
  <div class="callout-body">{@render children()}</div>
</div>
