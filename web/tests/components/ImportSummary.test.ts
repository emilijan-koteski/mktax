import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import ImportSummary from '../../src/components/ImportSummary.svelte';
import type { ParseResult } from '../../src/lib/types';
import parseGolden from '../fixtures/parse.json';

const info = parseGolden as ParseResult;

describe('ImportSummary', () => {
  it('lists the files with their row counts and lets the user remove one', async () => {
    const onremove = vi.fn();
    render(ImportSummary, { info, onremove });
    expect(screen.getByText('2024.csv')).toBeInTheDocument();
    expect(screen.getByText('8 реда')).toBeInTheDocument();
    const buttons = screen.getAllByRole('button', { name: /отстрани/i });
    expect(buttons).toHaveLength(3);
    await userEvent.click(buttons[0]);
    expect(onremove).toHaveBeenCalledWith('2024.csv');
  });

  it('shows the counts and the import flags in Macedonian', () => {
    render(ImportSummary, { info, onremove: vi.fn() });
    expect(screen.getByText(/10 трговски трансакции/)).toBeInTheDocument();
    expect(screen.getByText(/4 дивиденди/)).toBeInTheDocument();
    expect(screen.getByText(/3 камати/)).toBeInTheDocument();
    expect(screen.getByText(/бесплатни акции/)).toBeInTheDocument();
    expect(screen.getByText(/Mystery action/)).toBeInTheDocument();
  });
});
