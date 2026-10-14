import { render, screen, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import ReportView from '../../src/components/ReportView.svelte';
import type { CalculateResult } from '../../src/lib/types';
import reportGolden from '../fixtures/report.json';

const result = reportGolden as CalculateResult;

describe('ReportView', () => {
  beforeEach(() => {
    vi.stubGlobal('navigator', { ...navigator, clipboard: { writeText: vi.fn(async () => {}) } });
  });

  it('renders one section per month with entries and collapses empty ranges', () => {
    render(ReportView, { result, generatedAt: '2026-10-14T12:00:00Z' });
    const withEntries = result.report.months.filter((m) => m.has_entries).length;
    expect(screen.getAllByRole('region', { name: /2026|2025|2024/ })).toHaveLength(withEntries);
    expect(screen.getAllByText(/нема што да се пријави/).length).toBeGreaterThan(0);
    expect(screen.getAllByText('ДОСПЕАНО').length).toBe(withEntries);
    expect(screen.getByText('ВО ТЕК')).toBeInTheDocument();
  });

  it('shows the portal rows with the raw decimal behind the copy button', async () => {
    render(ReportView, { result, generatedAt: '2026-10-14T12:00:00Z' });
    const jan = screen.getByRole('region', { name: /јануари 2026/ });
    expect(within(jan).getByText('T8 Капитални добивки')).toBeInTheDocument();
    expect(within(jan).getByText(/S8\.1/)).toBeInTheDocument();
    expect(within(jan).getByText('САД')).toBeInTheDocument();
    const copy = within(jan).getByRole('button', { name: 'Копирај 42.50' });
    await userEvent.click(copy);
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('42.50');
    expect(within(jan).getByText(/Пријава до 10\.02\.2026/)).toBeInTheDocument();
  });

  it('never claims the report could not be computed because of unpublished future days', () => {
    render(ReportView, { result, generatedAt: '2026-10-14T12:00:00Z' });
    expect(screen.queryByText(/нема објавено курс/)).not.toBeInTheDocument();
  });

  it('marks a sale with unknown cost and lists open lots and excluded rows', () => {
    render(ReportView, { result, generatedAt: '2026-10-14T12:00:00Z' });
    const mar = screen.getByRole('region', { name: /март 2026/ });
    expect(within(mar).getByText(/продадено 6 повеќе/)).toBeInTheDocument();
    expect(within(screen.getByRole('region', { name: /Отворени лотови/ })).getByText('TEST_IE_ETF')).toBeInTheDocument();
    expect(screen.getByText(/Камата на готовина/)).toBeInTheDocument();
    expect(screen.getAllByText(/Return of capital/).length).toBeGreaterThan(0);
  });
});
