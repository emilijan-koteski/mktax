// @vitest-environment node
import { existsSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { runInThisContext } from 'node:vm';
import { describe, expect, it } from 'vitest';
import type { CalculateResult, Envelope, ParseResult } from '../src/lib/types';
import reportGolden from './fixtures/report.json';
import ratesGolden from './fixtures/rates.json';

const pub = join(import.meta.dirname, '..', 'public');
const wasmPath = join(pub, 'mktax.wasm');
const fixtures = join(import.meta.dirname, '..', '..', 'testdata', 't212');

type Engine = { parse(s: string): string; calculate(s: string): string; version(): string };

async function loadEngine(): Promise<Engine> {
  const g = globalThis as unknown as { Go?: new () => { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): Promise<void> }; mktax?: Engine };
  if (g.mktax) return g.mktax;
  runInThisContext(readFileSync(join(pub, 'wasm_exec.js'), 'utf8'));
  const go = new g.Go!();
  const { instance } = await WebAssembly.instantiate(readFileSync(wasmPath), go.importObject);
  void go.run(instance);
  for (let i = 0; !g.mktax && i < 100; i++) await new Promise((r) => setTimeout(r, 10));
  if (!g.mktax) throw new Error('engine did not register');
  return g.mktax;
}

describe.skipIf(!existsSync(wasmPath))('real mktax.wasm', () => {
  const files = ['2024.csv', '2025.csv', '2026.csv'].map((name) => ({ name, data: readFileSync(join(fixtures, name), 'utf8') }));

  it('parses the synthetic exports', async () => {
    const engine = await loadEngine();
    const out = JSON.parse(engine.parse(JSON.stringify({ files, now: '2026-10-14T12:00:00Z' }))) as Envelope<ParseResult>;
    expect(out.ok).toBe(true);
    if (out.ok) expect(out.result.rates_needed.months.length).toBe(12);
  }, 20_000);

  it('reproduces the Go golden report byte for byte', async () => {
    const engine = await loadEngine();
    const out = JSON.parse(engine.calculate(JSON.stringify({ files, rates: ratesGolden, now: '2026-10-14T12:00:00Z' }))) as Envelope<CalculateResult>;
    expect(out.ok).toBe(true);
    if (out.ok) expect(out.result).toEqual(reportGolden);
  }, 20_000);

  it('rejects non-string arguments', async () => {
    const engine = await loadEngine();
    expect(JSON.parse((engine.parse as unknown as (x: number) => string)(5)).error.code).toBe('bad_input');
  });
});
