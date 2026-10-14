<script lang="ts">
  import { fmtDateTime, fmtMonth } from '../lib/format';
  import { m, t } from '../lib/i18n';
  import type { Report } from '../lib/types';

  let { report, generatedAt }: { report: Report; generatedAt: string } = $props();
</script>

<header class="print-only border-b-2 border-ink pb-2">
  <h2 class="text-xl">{m.print.title}</h2>
  <p class="text-sm">{t(m.print.period, { from: fmtMonth(report.from), to: fmtMonth(report.to) })}</p>
  <p class="text-sm">{t(m.print.generated, { at: fmtDateTime(generatedAt) })}</p>
  <p class="text-sm">{t(m.print.files, { files: report.import.files.map((f) => f.name).join(', ') })}</p>
  <p class="text-sm">{m.print.method}</p>
  <p class="text-sm font-semibold">{m.print.notOfficial}</p>
</header>
