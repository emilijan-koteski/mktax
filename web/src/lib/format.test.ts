import { describe, expect, it } from 'vitest';
import { collator, fmtAmount, fmtDate, fmtDateTime, fmtMkd, fmtMonth, fmtQty, isNegative } from './format';

const norm = (s: string) => s.replace(/[\u00a0\u202f]/g, ' ');

describe('format', () => {
  it('formats MKD with a thousands dot, decimal comma and the ден. suffix', () => {
    expect(norm(fmtMkd('1234.5'))).toBe('1.234,50 ден.');
    expect(norm(fmtMkd('-1230'))).toBe('-1.230,00 ден.');
    expect(norm(fmtMkd('0'))).toBe('0,00 ден.');
  });

  it('formats foreign amounts with the currency code', () => {
    expect(norm(fmtAmount('42.5', 'USD'))).toBe('42,50 USD');
  });

  it('keeps quantities exact up to eight decimals', () => {
    expect(norm(fmtQty('0.12345678'))).toBe('0,12345678');
    expect(norm(fmtQty('6'))).toBe('6');
  });

  it('formats dates as dd.MM.yyyy and months in Macedonian', () => {
    expect(fmtDate('2026-05-10')).toBe('10.05.2026');
    expect(fmtDate('')).toBe('');
    expect(fmtMonth('2026-05')).toBe('мај 2026');
    expect(fmtMonth('2026-01')).toBe('јануари 2026');
    expect(fmtDateTime('2026-10-14T12:00:00Z')).toMatch(/^14\.10\.2026 \d{2}:\d{2}$/);
  });

  it('sorts Cyrillic names with the Macedonian collation', () => {
    expect(['Џ', 'А', 'Ѕ', 'Б'].sort(collator.compare)).toEqual(['А', 'Б', 'Ѕ', 'Џ']);
  });

  it('detects negative decimals without converting to numbers', () => {
    expect(isNegative('-0.01')).toBe(true);
    expect(isNegative('0.00')).toBe(false);
  });
});
