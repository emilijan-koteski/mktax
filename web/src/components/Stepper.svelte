<script lang="ts" module>
  export type StepState = 'done' | 'current' | 'todo';
</script>

<script lang="ts">
  import Check from '@lucide/svelte/icons/check';
  import { m } from '../lib/i18n';

  let { states }: { states: [StepState, StepState, StepState] } = $props();
  const labels = [m.steps.files, m.steps.review, m.steps.report];
</script>

<ol class="stepper" aria-label={m.steps.label}>
  {#each labels as label, i (label)}
    <li data-state={states[i]} aria-current={states[i] === 'current' ? 'step' : undefined}>
      <span class="step-num" data-state={states[i]}>
        {#if states[i] === 'done'}<Check strokeWidth={3} />{:else}{i + 1}{/if}
      </span>
      <span>{label}</span>
    </li>
  {/each}
</ol>
