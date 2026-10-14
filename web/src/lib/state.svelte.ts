import { ApiError, fetchRates } from './api';
import type { AppError, CalcOpts, CalculateResult, FileIn, ParseResult, RatesTable } from './types';
import { engine, loadEngine } from './wasm';

export type Phase = 'booting' | 'idle' | 'parsing' | 'parsed' | 'calculating' | 'report' | 'error';

const paint = () => new Promise<void>((r) => setTimeout(r, 0));

async function readText(f: File): Promise<string> {
  return typeof f.text === 'function' ? f.text() : new Response(f).text();
}

class AppState {
  phase = $state<Phase>('booting');
  error = $state<AppError | null>(null);
  files = $state.raw<FileIn[]>([]);
  importInfo = $state.raw<ParseResult | null>(null);
  fundUnits = $state<Record<string, boolean>>({});
  period = $state({ from: '', to: '' });
  report = $state.raw<CalculateResult | null>(null);
  reportOpts = $state.raw<CalcOpts | null>(null);
  ratesProgress = $state({ done: 0, total: 0 });
  generatedAt = $state<string | null>(null);
  printMode = $state(false);

  opts = $derived<CalcOpts>({ from: this.period.from, to: this.period.to, fund_units: $state.snapshot(this.fundUnits) });
  stale = $derived(this.phase === 'report' && JSON.stringify(this.opts) !== JSON.stringify(this.reportOpts));
  canCalculate = $derived((this.phase === 'parsed' || this.phase === 'report') && this.files.length > 0);

  async boot(): Promise<void> {
    try {
      await loadEngine();
      this.phase = 'idle';
    } catch (e) {
      this.fail({ code: 'engineLoad', message: String(e) });
    }
  }

  async addFiles(picked: Iterable<File>): Promise<void> {
    const incoming = await Promise.all([...picked].map(async (f) => ({ name: f.name, data: await readText(f) })));
    const byName: Record<string, FileIn> = Object.fromEntries(this.files.map((f) => [f.name, f]));
    for (const f of incoming) byName[f.name] = f;
    this.files = Object.values(byName);
    await this.parse();
  }

  removeFile(name: string): void {
    this.files = this.files.filter((f) => f.name !== name);
    if (this.files.length) void this.parse();
    else this.reset();
  }

  private async parse(): Promise<void> {
    this.phase = 'parsing';
    this.error = null;
    await paint();
    try {
      const res = engine.parse(this.files, $state.snapshot(this.fundUnits), new Date().toISOString());
      if (!res.ok) return this.fail(res.error);
      this.importInfo = res.result;
      this.fundUnits = Object.fromEntries(res.result.import.instruments.map((i) => [i.key, i.fund_units]));
      this.period = { ...res.result.suggested_period };
      this.report = null;
      this.reportOpts = null;
      this.phase = 'parsed';
    } catch (e) {
      this.fail(e);
    }
  }

  async calculate(): Promise<void> {
    if (!this.importInfo || !this.canCalculate) return;
    const opts = this.opts;
    this.phase = 'calculating';
    this.error = null;
    await paint();
    try {
      const need = this.importInfo.rates_needed;
      let rates: RatesTable = {};
      this.ratesProgress = { done: 0, total: need.months.length };
      if (need.months.length) {
        const fetched = await fetchRates(need.months, need.currencies, {
          onProgress: (done, total) => (this.ratesProgress = { done, total }),
        });
        rates = fetched.days;
      }
      const now = new Date().toISOString();
      const res = engine.calculate({ files: this.files, rates, from: opts.from, to: opts.to, fund_units: opts.fund_units, now });
      if (!res.ok) return this.fail(res.error);
      this.report = res.result;
      this.reportOpts = opts;
      this.generatedAt = now;
      this.phase = 'report';
    } catch (e) {
      this.fail(e);
    }
  }

  retry(): void {
    this.error = null;
    this.phase = this.importInfo ? 'parsed' : 'idle';
  }

  reset(): void {
    this.files = [];
    this.importInfo = null;
    this.report = null;
    this.reportOpts = null;
    this.fundUnits = {};
    this.period = { from: '', to: '' };
    this.error = null;
    this.phase = 'idle';
  }

  private fail(e: unknown): void {
    if (e instanceof ApiError) this.error = { code: e.code, message: e.message };
    else if (typeof e === 'object' && e !== null && 'code' in e) this.error = e as AppError;
    else this.error = { code: 'unknown', message: String(e) };
    this.phase = 'error';
  }
}

export const app = new AppState();
