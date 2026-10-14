import { afterEach, describe, expect, it, vi } from 'vitest';
import { ratesResponse } from '../../tests/fakeEngine';
import { ApiError, fetchRates, fetchRatesMonth } from './api';

afterEach(() => vi.unstubAllGlobals());

describe('rates api', () => {
  it('requests one relative URL per month and merges the days', async () => {
    const calls: string[] = [];
    vi.stubGlobal(
      'fetch',
      vi.fn(async (url: string) => {
        calls.push(url);
        const month = url.match(/rates\/(\d{4}-\d{2})/)![1];
        return new Response(JSON.stringify(ratesResponse(month, ['EUR', 'USD'])), { status: 200 });
      }),
    );
    const { days, missing } = await fetchRates(['2026-01', '2026-02'], ['EUR', 'USD']);
    expect(calls).toEqual(['/api/rates/2026-01?currencies=EUR%2CUSD', '/api/rates/2026-02?currencies=EUR%2CUSD']);
    expect(Object.keys(days).length).toBe(31 + 28);
    expect(days['2026-02-10'].USD.middle).toBe('50');
    expect(missing).toEqual([]);
  });

  it('collects missing days per month', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response(JSON.stringify({ ...ratesResponse('2026-10', ['EUR']), missing: ['2026-10-30'], complete: false }))),
    );
    const { missing } = await fetchRates(['2026-10'], ['EUR']);
    expect(missing).toEqual([{ month: '2026-10', days: ['2026-10-30'] }]);
  });

  it('throws an ApiError with the status and code on failure', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ error: { code: 'upstream', message: 'x' } }), { status: 502 })));
    await expect(fetchRatesMonth('2026-01', ['EUR'])).rejects.toMatchObject({ status: 502, code: 'upstream' });
    await expect(fetchRatesMonth('2026-01', ['EUR'])).rejects.toBeInstanceOf(ApiError);
  });

  it('reports network failures as ApiError without a status', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => { throw new TypeError('offline'); }));
    await expect(fetchRatesMonth('2026-01', ['EUR'])).rejects.toMatchObject({ status: 0, code: 'network' });
  });
});
