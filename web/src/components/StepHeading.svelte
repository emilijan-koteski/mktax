<script lang="ts">
  import Check from '@lucide/svelte/icons/check';
  import type { Snippet } from 'svelte';
  import type { StepState } from './Stepper.svelte';

  let { id, n, state, title, children }: { id: string; n: number; state: StepState; title: string; children?: Snippet } = $props();
</script>

<header class="flex flex-wrap items-center justify-between gap-x-4 gap-y-3">
  <div class="flex items-center gap-3">
    <span class="step-num" data-state={state} aria-hidden="true">
      {#if state === 'done'}<Check strokeWidth={3} />{:else}{n}{/if}
    </span>
    <h2 {id} class="text-xl"><span class="visually-hidden">{n}. </span>{title}</h2>
  </div>
  {@render children?.()}
</header>
