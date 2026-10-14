<script lang="ts">
  import { downloadText, portalRowsCsv } from '../lib/export/csv';
  import { fmtMonth } from '../lib/format';
  import { m, t } from '../lib/i18n';
  import { printTables } from '../lib/print';
  import { app } from '../lib/state.svelte';
  import type { CalculateResult, Month } from '../lib/types';
  import ExcludedRows from './ExcludedRows.svelte';
  import LedgerReport from './LedgerReport.svelte';
  import MonthCard from './MonthCard.svelte';
  import Notes from './Notes.svelte';
  import OpenLots from './OpenLots.svelte';
  import PrintHeader from './PrintHeader.svelte';
  import Warnings from './Warnings.svelte';

  let { result, generatedAt }: { result: CalculateResult; generatedAt: string } = $props();
  const report = $derived(result.report);

  type Group = { kind: 'month'; month: Month } | { kind: 'current'; month: Month } | { kind: 'empty'; from: string; to: string };

  const groups = $derived.by(() => {
    const out: Group[] = [];
    for (const month of report.months) {
      if (month.has_entries) out.push({ kind: 'month', month });
      else if (month.in_progress) out.push({ kind: 'current', month });
      else {
        const last = out.at(-1);
        if (last?.kind === 'empty') last.to = month.month;
        else out.push({ kind: 'empty', from: month.month, to: month.month });
      }
    }
    return out;
  });

  const printTitle = $derived(`${m.app.title} ${report.from} ${report.to}`);
  const tables = $derived(app.printMode && app.printLayout === 'tables');
  const csvName = $derived(t(m.csv.filename, { from: report.from, to: report.to }));

  function exportCsv() {
    downloadText(csvName, portalRowsCsv(report));
  }
</script>

<div>
  <div class="no-print flex flex-wrap items-center gap-2">
    <button type="button" class="btn" onclick={() => printTables(printTitle)}>{m.ledger.button}</button>
    <button type="button" class="btn btn-secondary" onclick={exportCsv}>{m.csv.download}</button>
  </div>

  {#if tables}<LedgerReport {report} {generatedAt} />{/if}

  <div class:no-print={tables}>
    <PrintHeader {report} {generatedAt} />

    <div class="mt-3"><Warnings warnings={report.warnings} raw={report.raw_warnings} /></div>

    {#each groups as g (g.kind === 'empty' ? `${g.from}-${g.to}` : g.month.month)}
      {#if g.kind === 'month'}
        <MonthCard month={g.month} printMode={app.printMode} />
      {:else if g.kind === 'current'}
        <p class="mt-6 flex flex-wrap items-baseline gap-x-3 border-t border-line pt-2 text-sm">
          <span class="font-display text-lg">{fmtMonth(g.month.month)}</span>
          <span class="badge badge-live">{m.month.status.inProgress}</span>
          <span class="text-muted">{m.month.nothingToFile}</span>
        </p>
      {:else}
        <p class="mt-6 border-t border-line pt-2 text-sm text-muted">
          {g.from === g.to ? `${fmtMonth(g.from)}: ${m.month.nothingToFile}` : t(m.month.emptyRange, { from: fmtMonth(g.from), to: fmtMonth(g.to) })}
        </p>
      {/if}
    {/each}

    <OpenLots positions={report.open_lots} asOf={generatedAt.slice(0, 10)} />
    <ExcludedRows summary={report.import} />
    <Notes notes={report.notes} />
  </div>
</div>
