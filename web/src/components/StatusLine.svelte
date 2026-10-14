<script lang="ts">
  import CircleAlert from '@lucide/svelte/icons/circle-alert';
  import CircleCheck from '@lucide/svelte/icons/circle-check';
  import LoaderCircle from '@lucide/svelte/icons/loader-circle';
  import { m, t } from '../lib/i18n';
  import { app } from '../lib/state.svelte';

  const text = $derived(
    app.phase === 'calculating' && app.ratesProgress.total > 0
      ? `${m.status.calculating} ${t(m.review.ratesProgress, app.ratesProgress)}`
      : m.status[app.phase],
  );
  const busy = $derived(app.phase === 'booting' || app.phase === 'parsing' || app.phase === 'calculating');
</script>

<p role="status" aria-live="polite" class="no-print text-caption text-balance text-ink-muted">
  {#if busy}
    <LoaderCircle class="mr-1.5 inline size-4 animate-spin align-[-3px] text-link" />
  {:else if app.phase === 'error'}
    <CircleAlert class="mr-1.5 inline size-4 align-[-3px] text-danger" />
  {:else}
    <CircleCheck class="mr-1.5 inline size-4 align-[-3px] text-success" />
  {/if}
  {text}
</p>
