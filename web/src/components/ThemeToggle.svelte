<script lang="ts">
  import Moon from '@lucide/svelte/icons/moon';
  import Sun from '@lucide/svelte/icons/sun';
  import { onMount } from 'svelte';
  import { m } from '../lib/i18n';
  import { applyTheme, currentTheme, storedTheme, type Theme } from '../lib/theme';

  let theme = $state<Theme>(currentTheme());
  const label = $derived(theme === 'dark' ? m.theme.toLight : m.theme.toDark);

  onMount(() => {
    const system = window.matchMedia?.('(prefers-color-scheme: dark)');
    const follow = (e: MediaQueryListEvent) => {
      if (storedTheme()) return;
      theme = e.matches ? 'dark' : 'light';
      applyTheme(theme, false);
    };
    system?.addEventListener('change', follow);
    return () => system?.removeEventListener('change', follow);
  });

  function toggle() {
    theme = theme === 'dark' ? 'light' : 'dark';
    applyTheme(theme, true);
  }
</script>

<button type="button" class="btn btn-ghost btn-icon" aria-label={label} title={label} onclick={toggle}>
  {#if theme === 'dark'}<Sun />{:else}<Moon />{/if}
</button>
