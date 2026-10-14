<script lang="ts">
  import { fmtAmount, fmtDate, fmtDateTime, fmtMonth, fmtNumber, fmtQty, isNegative, isZero, sumDecimals } from '../lib/format';
  import { m, t } from '../lib/i18n';
  import type { Lot, Report, Sale, Totals } from '../lib/types';
  import Warnings from './Warnings.svelte';

  let { report, generatedAt }: { report: Report; generatedAt: string } = $props();
  const c = m.ledger.cols;

  const months = $derived(report.months.filter((mo) => mo.has_entries));
  const sales = $derived(months.flatMap((mo) => mo.sales));
  const dividends = $derived(months.flatMap((mo) => mo.dividends));
  const warnings = $derived([...report.warnings, ...months.flatMap((mo) => mo.warnings)]);

  const single = (codes: string[]) => (new Set(codes).size === 1 ? codes[0] : '');
  const saleCur = $derived(single(sales.flatMap((s) => [s.currency, ...s.lots.map((l) => l.currency)])));
  const grossCur = $derived(single(dividends.map((d) => d.gross_currency)));
  const taxCur = $derived(single(dividends.map((d) => d.foreign_tax_currency)));
  const netCur = $derived(single(dividends.map((d) => d.net_currency)));
  const dividendCode = $derived([...new Set(dividends.map((d) => `${d.income_type} / ${d.income_subtype}`))].join(', '));

  const head = (label: string, cur: string) => (cur ? `${label} ${cur}` : label);
  const amount = (v: string, cur: string, shared: string) => (shared ? fmtNumber(v) : fmtAmount(v, cur));
  const sum = (values: string[]) => fmtNumber(sumDecimals(values));
  const total = (key: keyof Totals) => sum(months.map((mo) => String(mo.totals[key])));

  function status(lot: Lot): string {
    if (lot.status === 'taxable' && isNegative(lot.gain_mkd) && !isZero(lot.gain_mkd)) return m.ledger.loss;
    return m.lots.statuses[lot.status];
  }

  const security = (s: { name: string; ticker: string }) => s.name || s.ticker;
  const unknown = (lot: Lot) => lot.status === 'unknown_cost';
  const span = (s: Sale) => Math.max(1, s.lots.length);
</script>

<article class="ledger-doc print-only">
  <header>
    <h2>{m.ledger.title}</h2>
    <p>{t(m.ledger.account, { currency: report.currency, from: fmtMonth(report.from), to: fmtMonth(report.to) })}</p>
    <p>{t(m.ledger.source, { files: report.import.files.map((f) => f.name).join(', ') })}</p>
  </header>

  {#if warnings.length}<div class="mt-2"><Warnings {warnings} /></div>{/if}

  <section aria-labelledby="ledger-overview">
    <h3 id="ledger-overview">{m.ledger.overview}</h3>
    {#if months.length}
      <table class="grid-table">
        <thead>
          <tr>
            <th scope="col">{c.month}</th>
            <th scope="col">{c.fileBy}</th>
            <th scope="col">{c.payBy}</th>
            <th scope="col" class="num">{c.sales}</th>
            <th scope="col" class="num">{c.dividends}</th>
            <th scope="col" class="num">{c.taxableGain}</th>
            <th scope="col" class="num">{c.gainsTax}</th>
            <th scope="col" class="num">{c.dividendGross}</th>
            <th scope="col" class="num">{c.dividendTaxRejected}</th>
            <th scope="col" class="num">{c.dividendTaxAccepted}</th>
            <th scope="col" class="num">{c.dueRejected}</th>
            <th scope="col" class="num">{c.dueAccepted}</th>
          </tr>
        </thead>
        <tbody>
          {#each months as mo (mo.month)}
            <tr>
              <th scope="row">{fmtMonth(mo.month)}</th>
              <td>{fmtDate(mo.file_by)}</td>
              <td>{fmtDate(mo.pay_by)}</td>
              <td class="num">{mo.totals.sales}</td>
              <td class="num">{mo.totals.dividends}</td>
              <td class="num">{fmtNumber(mo.totals.taxable_gains_mkd)}</td>
              <td class="num">{fmtNumber(mo.totals.capital_gains_tax_mkd)}</td>
              <td class="num">{fmtNumber(mo.totals.dividend_gross_mkd)}</td>
              <td class="num">{fmtNumber(mo.totals.dividend_tax_if_credit_rejected_mkd)}</td>
              <td class="num">{fmtNumber(mo.totals.dividend_tax_if_credit_accepted_mkd)}</td>
              <td class="num">{fmtNumber(mo.totals.tax_due_if_credit_rejected_mkd)}</td>
              <td class="num">{fmtNumber(mo.totals.tax_due_if_credit_accepted_mkd)}</td>
            </tr>
          {/each}
        </tbody>
        <tfoot>
          <tr>
            <th scope="row" colspan="5">{m.common.total}</th>
            <td class="num">{total('taxable_gains_mkd')}</td>
            <td class="num">{total('capital_gains_tax_mkd')}</td>
            <td class="num">{total('dividend_gross_mkd')}</td>
            <td class="num">{total('dividend_tax_if_credit_rejected_mkd')}</td>
            <td class="num">{total('dividend_tax_if_credit_accepted_mkd')}</td>
            <td class="num">{total('tax_due_if_credit_rejected_mkd')}</td>
            <td class="num">{total('tax_due_if_credit_accepted_mkd')}</td>
          </tr>
        </tfoot>
      </table>
    {:else}
      <p>{m.ledger.none}</p>
    {/if}
  </section>

  {#if sales.length}
    <section aria-labelledby="ledger-sales">
      <h3 id="ledger-sales">{m.ledger.sales}</h3>
      <table class="grid-table">
        <thead>
          <tr>
            <th scope="col">{c.saleDate}</th>
            <th scope="col">{c.security}</th>
            <th scope="col" class="num">{c.quantity}</th>
            <th scope="col">{c.acquired}</th>
            <th scope="col" class="num">{head(c.cost, saleCur)}</th>
            <th scope="col" class="num">{c.acquisitionRate}</th>
            <th scope="col" class="num">{c.costMkd}</th>
            <th scope="col" class="num">{head(c.proceeds, saleCur)}</th>
            <th scope="col" class="num">{c.saleRate}</th>
            <th scope="col" class="num">{c.proceedsMkd}</th>
            <th scope="col" class="num">{c.gainMkd}</th>
            <th scope="col">{c.status}</th>
            <th scope="col" class="num">{c.taxableGainMkd}</th>
            <th scope="col" class="num">{c.taxMkd}</th>
          </tr>
        </thead>
        {#each sales as s, i (i)}
          <tbody class="sale">
            {#each s.lots.length ? s.lots : [null] as lot, j (j)}
              <tr>
                {#if j === 0}
                  <td rowspan={span(s)}>{fmtDate(s.date)}</td>
                  <th scope="row" rowspan={span(s)}>{security(s)}<span class="isin">{s.isin}</span></th>
                {/if}
                {#if lot}
                  <td class="num">{fmtQty(lot.quantity)}</td>
                  <td>{fmtDate(lot.acquired)}</td>
                  <td class="num">{unknown(lot) ? '?' : amount(lot.cost, lot.currency, saleCur)}</td>
                  <td class="num">{unknown(lot) ? '' : fmtNumber(lot.acquisition_rate, 0, 4)}</td>
                  <td class="num">{unknown(lot) ? '' : fmtNumber(lot.cost_mkd)}</td>
                  <td class="num">{amount(lot.proceeds, s.currency, saleCur)}</td>
                  <td class="num">{fmtNumber(s.sale_rate, 0, 4)}</td>
                  <td class="num">{fmtNumber(lot.proceeds_mkd)}</td>
                  <td class="num">{unknown(lot) ? '' : fmtNumber(lot.gain_mkd)}</td>
                  <td>{status(lot)}</td>
                {:else}
                  <td class="num">{fmtQty(s.quantity)}</td>
                  <td colspan="9"></td>
                {/if}
                {#if j === 0}
                  <td class="num" rowspan={span(s)}>{fmtNumber(s.taxable_gain_mkd)}</td>
                  <td class="num" rowspan={span(s)}>{fmtNumber(s.tax_mkd)}</td>
                {/if}
              </tr>
            {/each}
          </tbody>
        {/each}
        <tfoot>
          <tr>
            <th scope="row" colspan="12">{m.common.total}</th>
            <td class="num">{sum(sales.map((s) => s.taxable_gain_mkd))}</td>
            <td class="num">{sum(sales.map((s) => s.tax_mkd))}</td>
          </tr>
        </tfoot>
      </table>
    </section>
  {/if}

  {#if dividends.length}
    <section aria-labelledby="ledger-dividends">
      <h3 id="ledger-dividends">{t(m.ledger.dividends, { code: dividendCode })}</h3>
      <table class="grid-table">
        <thead>
          <tr>
            <th scope="col">{c.date}</th>
            <th scope="col">{c.payer}</th>
            <th scope="col">{c.country}</th>
            <th scope="col" class="num">{head(c.gross, grossCur)}</th>
            <th scope="col" class="num">{head(c.withheld, taxCur)}</th>
            <th scope="col" class="num">{head(c.net, netCur)}</th>
            <th scope="col" class="num">{c.rate}</th>
            <th scope="col" class="num">{c.grossMkd}</th>
            <th scope="col" class="num">{c.withheldMkd}</th>
            <th scope="col" class="num">{c.taxRejected}</th>
            <th scope="col" class="num">{c.taxAccepted}</th>
          </tr>
        </thead>
        <tbody>
          {#each dividends as d, i (i)}
            <tr>
              <td>{fmtDate(d.date)}</td>
              <th scope="row">{security(d)}<span class="isin">{d.isin}</span></th>
              <td>{d.country.name_mk}</td>
              <td class="num">{amount(d.gross, d.gross_currency, grossCur)}</td>
              <td class="num">{amount(d.foreign_tax, d.foreign_tax_currency, taxCur)}</td>
              <td class="num">{amount(d.net, d.net_currency, netCur)}</td>
              <td class="num">{fmtNumber(d.gross_rate, 0, 4)}</td>
              <td class="num">{fmtNumber(d.gross_mkd)}</td>
              <td class="num">{fmtNumber(d.withheld_mkd)}</td>
              <td class="num">{fmtNumber(d.tax_if_credit_rejected_mkd)}</td>
              <td class="num">{fmtNumber(d.tax_if_credit_accepted_mkd)}</td>
            </tr>
          {/each}
        </tbody>
        <tfoot>
          <tr>
            <th scope="row" colspan="7">{m.common.total}</th>
            <td class="num">{sum(dividends.map((d) => d.gross_mkd))}</td>
            <td class="num">{sum(dividends.map((d) => d.withheld_mkd))}</td>
            <td class="num">{sum(dividends.map((d) => d.tax_if_credit_rejected_mkd))}</td>
            <td class="num">{sum(dividends.map((d) => d.tax_if_credit_accepted_mkd))}</td>
          </tr>
        </tfoot>
      </table>
    </section>
  {/if}

  <section aria-labelledby="ledger-method">
    <h3 id="ledger-method">{m.ledger.method}</h3>
    <ul>
      {#each [m.ledger.rules.fifo, m.ledger.rules.amounts, m.ledger.rules.fx, m.ledger.rules.tax, m.ledger.rules.dividends] as rule (rule)}
        <li>{rule}</li>
      {/each}
      <li><strong>{t(m.ledger.rules.notOfficial, { at: fmtDateTime(generatedAt) })}</strong></li>
    </ul>
  </section>
</article>
