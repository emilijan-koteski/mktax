<script lang="ts">
  import { m, t } from '../lib/i18n';
  import { app } from '../lib/state.svelte';

  const text = $derived(
    app.phase === 'calculating' && app.ratesProgress.total > 0
      ? `${m.status.calculating} ${t(m.review.ratesProgress, app.ratesProgress)}`
      : m.status[app.phase],
  );
</script>

<p role="status" aria-live="polite" class="no-print text-sm text-muted">{text}</p>
