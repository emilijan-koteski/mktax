<script lang="ts">
  import { onMount } from 'svelte';
  import { version } from '../package.json';
  import Callout from './components/Callout.svelte';
  import Disclaimer from './components/Disclaimer.svelte';
  import FileDrop from './components/FileDrop.svelte';
  import Help from './components/Help.svelte';
  import ImportSummary from './components/ImportSummary.svelte';
  import InstrumentsTable from './components/InstrumentsTable.svelte';
  import PeriodPicker from './components/PeriodPicker.svelte';
  import Privacy from './components/Privacy.svelte';
  import ReportView from './components/ReportView.svelte';
  import StatusLine from './components/StatusLine.svelte';
  import { m, t } from './lib/i18n';
  import { app } from './lib/state.svelte';
  import { errorText } from './lib/warnings';

  const months = $derived.by(() => {
    const p = app.importInfo?.suggested_period;
    if (!p) return [];
    const out: string[] = [];
    for (let [y, mo] = p.from.split('-').map(Number); `${y}-${String(mo).padStart(2, '0')}` <= p.to; mo === 12 ? ((y += 1), (mo = 1)) : (mo += 1)) {
      out.push(`${y}-${String(mo).padStart(2, '0')}`);
    }
    return out;
  });
  const busy = $derived(app.phase === 'parsing' || app.phase === 'calculating');

  onMount(() => {
    void app.boot();
    const before = () => (app.printMode = true);
    const after = () => (app.printMode = false);
    window.addEventListener('beforeprint', before);
    window.addEventListener('afterprint', after);
    return () => {
      window.removeEventListener('beforeprint', before);
      window.removeEventListener('afterprint', after);
    };
  });
</script>

<div class="mx-auto max-w-6xl px-4 pb-16 pt-6">
  <header class="no-print flex flex-wrap items-baseline justify-between gap-x-6 gap-y-2">
    <div>
      <h1 class="text-4xl">{m.app.title}</h1>
      <p class="text-muted">{m.app.subtitle}</p>
    </div>
    <nav class="flex gap-4 text-sm">
      <a href="#help" class="text-accent underline">{m.app.help}</a>
      <a href="#privacy" class="text-accent underline">{m.app.privacy}</a>
    </nav>
  </header>

  <div class="mt-4"><Disclaimer /></div>
  <div class="mt-3"><StatusLine /></div>

  {#if app.error}
    <div class="no-print mt-3">
      <Callout kind="error">
        {errorText(app.error)}
        <button type="button" class="btn btn-secondary ml-2" onclick={() => app.retry()}>{m.errors.retry}</button>
      </Callout>
    </div>
  {/if}

  <main>
    <section class="no-print mt-8" aria-labelledby="step-files">
      <h2 id="step-files" class="text-2xl">1. {m.steps.files}</h2>
      <div class="mt-3"><FileDrop onfiles={(files) => void app.addFiles(files)} {busy} /></div>
    </section>

    {#if app.importInfo}
      <section class="no-print mt-8" aria-labelledby="step-review">
        <h2 id="step-review" class="text-2xl">2. {m.steps.review}</h2>
        <div class="mt-3"><ImportSummary info={app.importInfo} onremove={(name) => app.removeFile(name)} /></div>
        <h3 class="mt-4 text-lg">{m.review.instruments}</h3>
        <p class="max-w-prose text-sm text-muted">{m.review.etfHint}</p>
        <div class="mt-2"><InstrumentsTable instruments={app.importInfo.import.instruments} bind:fundUnits={app.fundUnits} /></div>
        <div class="mt-4 flex flex-wrap items-center gap-4">
          <PeriodPicker {months} bind:period={app.period} />
          <button type="button" class="btn" disabled={!app.canCalculate || busy} onclick={() => void app.calculate()}>{m.review.calculate}</button>
        </div>
        {#if app.stale}
          <div class="mt-3">
            <Callout kind="info">
              {m.review.stale}
              <button type="button" class="btn ml-2" onclick={() => void app.calculate()}>{m.review.recalculate}</button>
            </Callout>
          </div>
        {/if}
      </section>
    {/if}

    {#if app.report && app.generatedAt}
      <section class="mt-8" aria-labelledby="step-report" class:opacity-60={app.stale}>
        <h2 id="step-report" class="no-print text-2xl">3. {m.steps.report}</h2>
        <div class="mt-3"><ReportView result={app.report} generatedAt={app.generatedAt} /></div>
      </section>
    {/if}
  </main>

  <footer class="no-print mt-12 space-y-3 border-t border-line pt-4 text-sm">
    <Help />
    <Privacy />
    <p class="text-xs text-muted">{t(m.app.version, { version })}</p>
  </footer>
</div>
