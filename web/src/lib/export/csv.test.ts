import { describe, expect, it } from 'vitest';
import type { CalculateResult } from '../types';
import { portalRowsCsv } from './csv';
import golden from '../../../tests/fixtures/report.json';

const result = golden as CalculateResult;

describe('portalRowsCsv', () => {
  const csv = portalRowsCsv(result.report);
  const lines = csv.split('\r\n');

  it('starts with a BOM and uses CRLF with semicolons', () => {
    expect(csv.startsWith('﻿')).toBe(true);
    expect(lines[0].split(';').length).toBeGreaterThanOrEqual(18);
    expect(csv.endsWith('\r\n')).toBe(true);
  });

  it('writes one line per portal row', () => {
    const rows = result.report.months.reduce((n, m) => n + m.portal_rows.length, 0);
    expect(lines.length - 2).toBe(rows);
  });

  it('uses a decimal comma in numeric columns and keeps the raw ISO date', () => {
    const jan = lines.find((l) => l.includes('2026-01-15'));
    expect(jan).toContain('42,50');
    expect(jan).toContain('2026-01;2026-02-10;2026-02-15');
  });

  it('quotes fields containing the delimiter or quotes', () => {
    const copy = structuredClone(result.report);
    const sample = result.report.months.find((m) => m.portal_rows.length)!.portal_rows[0];
    copy.months[0].portal_rows = [{ ...sample, name: 'A; "B"' }];
    const out = portalRowsCsv(copy);
    expect(out).toContain('"A; ""B"""');
  });
});
