<script lang="ts">
  import Check from '@lucide/svelte/icons/check';
  import Copy from '@lucide/svelte/icons/copy';
  import { m, t } from '../lib/i18n';

  let { value, display = value }: { value: string; display?: string } = $props();
  let copied = $state(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(value);
      copied = true;
      setTimeout(() => (copied = false), 1500);
    } catch {
      window.prompt(m.portal.copy.replace(' {value}', ''), value);
    }
  }
</script>

<span class="copy">
  <span class="copy-value">{display}</span>
  <button
    type="button"
    class="btn btn-ghost btn-icon no-print"
    data-copied={copied || undefined}
    aria-label={t(m.portal.copy, { value })}
    title={copied ? m.portal.copied : t(m.portal.copy, { value })}
    onclick={copy}
  >
    {#if copied}<Check />{:else}<Copy />{/if}
  </button>
</span>
