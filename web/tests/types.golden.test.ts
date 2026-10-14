import { describe, expect, it } from 'vitest';
import type { CalculateResult, ParseResult, RatesTable, Warning } from '../src/lib/types';
import parseGolden from './fixtures/parse.json';
import ratesGolden from './fixtures/rates.json';
import reportGolden from './fixtures/report.json';

type Shape<T> = T extends Warning ? { code: string; params: object } : T extends string ? string : T extends (infer U)[] ? Shape<U>[] : T extends object ? { [K in keyof T]: Shape<T[K]> } : T;

const parsed = parseGolden satisfies Shape<ParseResult>;
const calculated = reportGolden satisfies Shape<CalculateResult>;
const rates = ratesGolden satisfies Shape<RatesTable>;

describe('golden fixtures match the TypeScript types', () => {
  it('keeps the shapes the Go side emits', () => {
    expect(parsed.rates_needed.months.length).toBe(12);
    expect(calculated.report.months.length).toBe(34);
    expect(Object.keys(rates).length).toBeGreaterThan(300);
    expect(calculated.report.months[28].portal_rows[0].kind).toBe('dividend');
  });
});
