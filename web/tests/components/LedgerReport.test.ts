import { render, screen, within } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import LedgerReport from '../../src/components/LedgerReport.svelte';
import { fmtNumber, sumDecimals } from '../../src/lib/format';
import type { CalculateResult } from '../../src/lib/types';
import reportGolden from '../fixtures/report.json';

const report = (reportGolden as CalculateResult).report;
const months = report.months.filter((m) => m.has_entries);

describe('LedgerReport', () => {
  it('lists one overview row per month with entries and an exact total', () => {
    render(LedgerReport, { report, generatedAt: '2026-10-14T12:00:00Z' });
    const overview = screen.getByRole('region', { name: /Преглед по месеци/ });
    expect(within(overview).getAllByRole('row')).toHaveLength(months.length + 2);
    const total = fmtNumber(sumDecimals(months.map((m) => m.totals.tax_due_if_credit_accepted_mkd)));
    const footer = within(overview).getAllByRole('row').at(-1)!;
    expect(within(footer).getAllByRole('cell').at(-1)).toHaveTextContent(total);
  });

  it('prints one row per FIFO lot and one row per dividend', () => {
    render(LedgerReport, { report, generatedAt: '2026-10-14T12:00:00Z' });
    const sales = months.flatMap((m) => m.sales);
    const lots = sales.reduce((n, s) => n + Math.max(1, s.lots.length), 0);
    expect(within(screen.getByRole('region', { name: /FIFO/ })).getAllByRole('row')).toHaveLength(lots + 2);
    const dividends = months.flatMap((m) => m.dividends);
    expect(within(screen.getByRole('region', { name: /Дивиденди/ })).getAllByRole('row')).toHaveLength(dividends.length + 2);
    expect(screen.getByText(/Начин на пресметка/)).toBeInTheDocument();
  });
});
