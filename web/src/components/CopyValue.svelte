<script lang="ts">
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

<span class="inline-flex items-center gap-1">
  <span class="num font-medium">{display}</span>
  <button
    type="button"
    class="no-print rounded-xs border border-line px-1 text-xs text-muted hover:border-accent hover:text-accent"
    aria-label={t(m.portal.copy, { value })}
    onclick={copy}
  >
    {copied ? m.portal.copied : '⧉'}
  </button>
</span>
