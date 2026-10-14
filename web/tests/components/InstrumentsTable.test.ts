import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import InstrumentsTable from '../../src/components/InstrumentsTable.svelte';
import type { ParseResult } from '../../src/lib/types';
import parseGolden from '../fixtures/parse.json';

const info = parseGolden as ParseResult;

describe('InstrumentsTable', () => {
  it('renders one ETF checkbox per instrument with the guessed state', () => {
    const fundUnits = Object.fromEntries(info.import.instruments.map((i) => [i.key, i.fund_units]));
    render(InstrumentsTable, { instruments: info.import.instruments, fundUnits });
    const boxes = screen.getAllByRole('checkbox') as HTMLInputElement[];
    expect(boxes).toHaveLength(3);
    expect(boxes.map((b) => b.checked)).toEqual([false, true, false]);
    expect(screen.getByText('TEST_IE_ETF')).toBeInTheDocument();
    expect(screen.getAllByText(/претпоставено/).length).toBeGreaterThan(0);
  });
});
