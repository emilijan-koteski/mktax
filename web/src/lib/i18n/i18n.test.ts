import { describe, expect, it } from 'vitest';
import { keyPaths, m, plural, t } from './index';
import { mk } from './mk';

describe('i18n', () => {
  it('has a non-empty string at every leaf', () => {
    const paths = keyPaths(mk);
    expect(paths.length).toBeGreaterThan(50);
    for (const p of paths) {
      const value = p.split('.').reduce<unknown>((o, k) => (o as Record<string, unknown>)[k], mk as unknown);
      expect(typeof value, p).toBe('string');
      expect((value as string).trim().length, p).toBeGreaterThan(0);
      expect(value as string, p).not.toMatch(/TODO/);
    }
    expect(new Set(paths).size).toBe(paths.length);
  });

  it('exposes the Macedonian catalog as the current one', () => {
    expect(m.app.title).toBe('Даночко');
  });

  it('interpolates named parameters and leaves unknown ones visible', () => {
    expect(t('{a} и {b}', { a: 1, b: 'x' })).toBe('1 и x');
    expect(t('{missing}', {})).toBe('{missing}');
  });

  it('picks Macedonian plural forms', () => {
    const forms = { one: '{n} ставка', other: '{n} ставки' };
    expect(plural(1, forms)).toBe('1 ставка');
    expect(plural(21, forms)).toBe('21 ставка');
    expect(plural(5, forms)).toBe('5 ставки');
  });
});
