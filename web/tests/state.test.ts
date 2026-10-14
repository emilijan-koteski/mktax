import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { installFakeEngine, ratesResponse } from './fakeEngine';
import parseGolden from './fixtures/parse.json';

async function freshApp() {
  vi.resetModules();
  const mod = await import('../src/lib/state.svelte');
  return mod.app;
}

function csvFile(name = '2025.csv') {
  return new File(['Action,Time (UTC)\n'], name, { type: 'text/csv' });
}

describe('app state', () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  beforeEach(() => {
    installFakeEngine();
    fetchMock = vi.fn(async (url: string) => {
      const month = url.match(/rates\/(\d{4}-\d{2})/)![1];
      return new Response(JSON.stringify(ratesResponse(month, ['EUR', 'USD'])));
    });
    vi.stubGlobal('fetch', fetchMock);
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date('2026-10-14T12:00:00Z'));
  });
  afterEach(() => {
    vi.doUnmock('../src/lib/wasm');
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it('boots to idle when the engine is present', async () => {
    const app = await freshApp();
    expect(app.phase).toBe('booting');
    await app.boot();
    expect(app.phase).toBe('idle');
  });

  it('parses files added while the engine is still loading once it is ready', async () => {
    let release!: () => void;
    const gate = new Promise<void>((r) => (release = r));
    vi.doMock('../src/lib/wasm', async (load) => ({ ...(await load<typeof import('../src/lib/wasm')>()), loadEngine: () => gate }));
    const app = await freshApp();
    const booting = app.boot();
    const adding = app.addFiles([csvFile()]);
    await new Promise((r) => setTimeout(r, 0));
    expect(app.phase).toBe('booting');
    release();
    await booting;
    await adding;
    expect(app.phase).toBe('parsed');
    expect(app.error).toBeNull();
  });

  it('parses added files and seeds the options from the import', async () => {
    const app = await freshApp();
    await app.boot();
    await app.addFiles([csvFile()]);
    expect(app.phase).toBe('parsed');
    expect(app.importInfo?.import.counts.trades).toBe(10);
    expect(app.fundUnits).toEqual({ GB0000000003: false, IE0000000002: true, US0000000001: false });
    expect(app.period).toEqual({ from: '2024-01', to: '2026-10' });
    const parse = (globalThis as unknown as { mktax: { parse: ReturnType<typeof vi.fn> } }).mktax.parse;
    const sent = JSON.parse(parse.mock.calls[0][0] as string);
    expect(sent.files[0].name).toBe('2025.csv');
    expect(sent.now).toBe('2026-10-14T12:00:00.000Z');
  });

  it('fetches one month at a time with relative URLs and stores the report', async () => {
    const app = await freshApp();
    await app.boot();
    await app.addFiles([csvFile()]);
    await app.calculate();
    expect(app.phase).toBe('report');
    expect(fetchMock).toHaveBeenCalledTimes(parseGolden.rates_needed.months.length);
    expect(fetchMock.mock.calls[0][0]).toBe('/api/rates/2024-01?currencies=EUR%2CUSD');
    expect(app.report?.report.months.length).toBe(34);
    expect(app.stale).toBe(false);
    app.fundUnits.IE0000000002 = false;
    expect(app.stale).toBe(true);
  });

  it('skips the network when no rates are needed', async () => {
    installFakeEngine({ parse: { ok: true, result: { ...parseGolden, rates_needed: { months: [], currencies: [] } } } });
    const app = await freshApp();
    await app.boot();
    await app.addFiles([csvFile()]);
    await app.calculate();
    expect(fetchMock).not.toHaveBeenCalled();
    expect(app.phase).toBe('report');
  });

  it('surfaces engine errors and can retry', async () => {
    installFakeEngine({ calculate: { ok: false, error: { code: 'rates_missing', message: 'x', missing: [{ currency: 'EUR', date: '2026-01-15' }] } } });
    const app = await freshApp();
    await app.boot();
    await app.addFiles([csvFile()]);
    await app.calculate();
    expect(app.phase).toBe('error');
    expect(app.error?.code).toBe('rates_missing');
    app.retry();
    expect(app.phase).toBe('parsed');
    expect(app.error).toBeNull();
  });

  it('turns a failing rates request into a network error', async () => {
    fetchMock.mockImplementation(async () => new Response('{"error":{"code":"upstream","message":"x"}}', { status: 502 }));
    const app = await freshApp();
    await app.boot();
    await app.addFiles([csvFile()]);
    await app.calculate();
    expect(app.phase).toBe('error');
    expect(app.error?.code).toBe('upstream');
  });

  it('replaces a file with the same name and clears everything on reset', async () => {
    const app = await freshApp();
    await app.boot();
    await app.addFiles([csvFile('a.csv')]);
    await app.addFiles([csvFile('a.csv'), csvFile('b.csv')]);
    expect(app.files.map((f) => f.name)).toEqual(['a.csv', 'b.csv']);
    app.removeFile('b.csv');
    expect(app.files.length).toBe(1);
    app.reset();
    expect(app.phase).toBe('idle');
    expect(app.files).toEqual([]);
    expect(app.report).toBeNull();
  });
});
