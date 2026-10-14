import { vi } from 'vitest';
import parseGolden from './fixtures/parse.json';
import reportGolden from './fixtures/report.json';

export function installFakeEngine(overrides: { parse?: unknown; calculate?: unknown } = {}) {
  const parse = vi.fn(() => JSON.stringify(overrides.parse ?? { ok: true, result: parseGolden }));
  const calculate = vi.fn(() => JSON.stringify(overrides.calculate ?? { ok: true, result: reportGolden }));
  const version = vi.fn(() => 'test');
  vi.stubGlobal('mktax', { parse, calculate, version });
  vi.stubGlobal('Go', class {});
  return { parse, calculate };
}

export function ratesResponse(month: string, currencies: string[]) {
  const [y, m] = month.split('-').map(Number);
  const days: Record<string, Record<string, { units: number; middle: string }>> = {};
  for (let d = new Date(Date.UTC(y, m - 1, 1)); d.getUTCMonth() === m - 1; d.setUTCDate(d.getUTCDate() + 1)) {
    const key = d.toISOString().slice(0, 10);
    days[key] = Object.fromEntries(currencies.map((c) => [c, { units: 1, middle: '50' }]));
  }
  return { month, from: `${month}-01`, to: Object.keys(days).at(-1), days, missing: [], complete: true };
}
