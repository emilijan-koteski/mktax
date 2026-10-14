import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import MonthCard from '../../src/components/MonthCard.svelte';
import type { CalculateResult } from '../../src/lib/types';
import reportGolden from '../fixtures/report.json';

const months = (reportGolden as CalculateResult).report.months;
const jan = months.find((m) => m.month === '2026-01')!;
const oct = months.find((m) => m.month === '2026-10')!;

describe('MonthCard', () => {
  it('shows the deadlines, the tax summary and the details on demand', async () => {
    render(MonthCard, { month: jan, printMode: false });
    expect(screen.getByText(/Пријава до 10\.02\.2026/)).toBeInTheDocument();
    expect(screen.getByText(/данок за плаќање: 235,24/)).toBeInTheDocument();
    expect(screen.queryByText(/Лотови \(FIFO\)/)).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: /детали/i }));
    expect(screen.getByText(/Лотови \(FIFO\)/)).toBeInTheDocument();
    expect(screen.getByText(/ослободено \(над 2 години\)/)).toBeInTheDocument();
  });

  it('forces the details open in print mode and labels the current month', () => {
    render(MonthCard, { month: oct, printMode: true });
    expect(screen.getByText('ВО ТЕК')).toBeInTheDocument();
    expect(screen.getByText(/нема што да се пријави/)).toBeInTheDocument();
  });
});
