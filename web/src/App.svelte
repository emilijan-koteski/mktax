<script lang="ts">
  import Calculator from '@lucide/svelte/icons/calculator';
  import CircleHelp from '@lucide/svelte/icons/circle-help';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import ShieldCheck from '@lucide/svelte/icons/shield-check';
  import { onMount } from 'svelte';
  import { version } from '../package.json';
  import coffee from './assets/bmc-cup.gif';
  import logo from './assets/danocko-logo.png';
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
  import StepHeading from './components/StepHeading.svelte';
  import Stepper, { type StepState } from './components/Stepper.svelte';
  import ThemeToggle from './components/ThemeToggle.svelte';
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
  const steps = $derived<[StepState, StepState, StepState]>(
    app.report ? ['done', 'done', 'current'] : app.importInfo ? ['done', 'current', 'todo'] : ['current', 'todo', 'todo'],
  );

  function reveal(id: string) {
    const el = document.getElementById(id);
    if (el instanceof HTMLDetailsElement) el.open = true;
  }

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

<header class="no-print sticky top-0 z-20 border-b border-line bg-surface">
  <div class="mx-auto flex h-16 max-w-6xl items-center justify-between gap-2 px-4 sm:gap-3 sm:px-6">
    <a href="#top" class="flex min-w-0 items-center gap-2 text-ink no-underline sm:gap-3">
      <img src={logo} alt="" width="40" height="40" class="size-9 shrink-0 sm:size-10" />
      <div class="min-w-0">
        <h1 class="text-xl leading-6 font-bold tracking-tight">{m.app.title}</h1>
        <p class="hidden truncate text-xs text-ink-muted sm:block">{m.app.subtitle}</p>
      </div>
    </a>
    <nav class="flex shrink-0 items-center sm:gap-1">
      <a href="#help" class="btn btn-ghost btn-sm px-2 sm:px-3" onclick={() => reveal('help')}><CircleHelp class="hidden sm:block" />{m.app.help}</a>
      <a href="#privacy" class="btn btn-ghost btn-sm px-2 sm:px-3" onclick={() => reveal('privacy')}><ShieldCheck class="hidden sm:block" />{m.app.privacy}</a>
      <ThemeToggle />
    </nav>
  </div>
</header>

<div id="top" class="mx-auto max-w-6xl px-4 pt-6 pb-16 sm:px-6 sm:pt-10">
  <div class="no-print mx-auto max-w-2xl space-y-6 text-center">
    <p class="text-lg text-balance text-ink-muted">{m.app.lead}</p>
    <Stepper states={steps} />
    <StatusLine />
  </div>

  <div class="mt-6 space-y-3">
    <Disclaimer />
    {#if app.error}
      <div class="no-print">
        <Callout kind="error">
          <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
            <span class="flex-1">{errorText(app.error)}</span>
            <button type="button" class="btn btn-secondary btn-sm" onclick={() => app.retry()}><RefreshCw />{m.errors.retry}</button>
          </div>
        </Callout>
      </div>
    {/if}
  </div>

  <main class="mt-8 space-y-10">
    <section class="no-print card space-y-5 scroll-mt-24" aria-labelledby="step-files">
      <StepHeading id="step-files" n={1} state={steps[0]} title={m.steps.files} />
      <FileDrop onfiles={(files) => void app.addFiles(files)} {busy} />
      {#if app.importInfo}
        <div class="border-t border-line pt-5">
          <ImportSummary info={app.importInfo} onremove={(name) => app.removeFile(name)} />
        </div>
      {/if}
    </section>

    {#if app.importInfo}
      <section class="no-print card space-y-5 scroll-mt-24" aria-labelledby="step-review">
        <StepHeading id="step-review" n={2} state={steps[1]} title={m.steps.review} />
        <div class="space-y-2">
          <h3 class="card-title">{m.review.instruments}</h3>
          <p class="max-w-prose text-caption text-ink-muted">{m.review.etfHint}</p>
        </div>
        <InstrumentsTable instruments={app.importInfo.import.instruments} bind:fundUnits={app.fundUnits} />
        <div class="flex flex-col gap-4 border-t border-line pt-5 sm:flex-row sm:items-end sm:justify-between">
          <PeriodPicker {months} bind:period={app.period} />
          <button type="button" class="btn btn-primary w-full sm:w-auto" disabled={!app.canCalculate || busy} onclick={() => void app.calculate()}>
            <Calculator />{m.review.calculate}
          </button>
        </div>
        {#if app.stale}
          <Callout kind="info">
            <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
              <span class="flex-1">{m.review.stale}</span>
              <button type="button" class="btn btn-primary btn-sm" onclick={() => void app.calculate()}><RefreshCw />{m.review.recalculate}</button>
            </div>
          </Callout>
        {/if}
      </section>
    {/if}

    {#if app.report && app.generatedAt}
      <section class="space-y-5 scroll-mt-24 transition-opacity" aria-labelledby="step-report" class:opacity-60={app.stale}>
        <div class="no-print"><StepHeading id="step-report" n={3} state={steps[2]} title={m.steps.report} /></div>
        <ReportView result={app.report} generatedAt={app.generatedAt} />
      </section>
    {/if}
  </main>

  <footer class="no-print mt-16 space-y-4">
    <div class="grid items-start gap-4 md:grid-cols-2">
      <Help />
      <Privacy />
    </div>
    <div class="space-y-1.5 text-center text-xs text-ink-muted">
      <p class="flex items-center justify-center gap-2">
        <img src={logo} alt="" width="20" height="20" class="size-5" />{m.app.title} · {t(m.app.version, { version })}
      </p>
      <p>{m.footer.madeBy} <a href="mailto:emilijankoteski@pm.me" title="emilijankoteski@pm.me">Emilijan Koteski</a></p>
      <p>
        {m.footer.hobby} <a href="https://buymeacoffee.com/emilijan" target="_blank" rel="noopener noreferrer">{m.footer.coffee}</a>.
        <img src={coffee} alt="" aria-hidden="true" width="18" height="18" class="inline-block size-[1.35em] align-[-0.3em]" />
      </p>
    </div>
  </footer>
</div>
