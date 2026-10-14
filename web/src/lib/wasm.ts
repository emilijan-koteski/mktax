import type { CalculateInput, CalculateResult, Envelope, FileIn, ParseResult } from './types';

let loading: Promise<void> | undefined;

export function loadEngine(url = '/mktax.wasm'): Promise<void> {
  if (globalThis.mktax) return Promise.resolve();
  loading ??= (async () => {
    if (typeof globalThis.Go !== 'function') throw new Error('wasm_exec.js is not loaded');
    const go = new globalThis.Go();
    const resp = await fetch(url);
    if (!resp.ok) throw new Error(`mktax.wasm: HTTP ${resp.status}`);
    const streaming = 'instantiateStreaming' in WebAssembly && (resp.headers.get('content-type') ?? '').startsWith('application/wasm');
    const { instance } = streaming
      ? await WebAssembly.instantiateStreaming(resp, go.importObject)
      : await WebAssembly.instantiate(await resp.arrayBuffer(), go.importObject);
    void go.run(instance);
    for (let i = 0; !globalThis.mktax && i < 100; i++) await new Promise((r) => setTimeout(r, 20));
    if (!globalThis.mktax) throw new Error('the engine did not register globalThis.mktax');
  })();
  return loading;
}

function eng() {
  if (!globalThis.mktax) throw new Error('engine not loaded');
  return globalThis.mktax;
}

export const engine = {
  parse(files: FileIn[], fundUnits: Record<string, boolean>, now: string): Envelope<ParseResult> {
    return JSON.parse(eng().parse(JSON.stringify({ files, fund_units: fundUnits, now }))) as Envelope<ParseResult>;
  },
  calculate(input: CalculateInput): Envelope<CalculateResult> {
    return JSON.parse(eng().calculate(JSON.stringify(input))) as Envelope<CalculateResult>;
  },
  version(): string {
    return eng().version();
  },
};
