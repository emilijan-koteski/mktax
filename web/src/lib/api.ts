import type { RatesMonthResponse, RatesTable, YearMonth } from './types';

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export async function fetchRatesMonth(month: YearMonth, currencies: string[], signal?: AbortSignal): Promise<RatesMonthResponse> {
  const q = new URLSearchParams({ currencies: currencies.join(',') });
  let resp: Response;
  try {
    resp = await fetch(`/api/rates/${month}?${q}`, { signal, headers: { Accept: 'application/json' } });
  } catch (e) {
    throw new ApiError(0, 'network', String(e));
  }
  if (!resp.ok) {
    const body = (await resp.json().catch(() => null)) as { error?: { code?: string; message?: string } } | null;
    throw new ApiError(resp.status, body?.error?.code ?? `http_${resp.status}`, body?.error?.message ?? resp.statusText);
  }
  return (await resp.json()) as RatesMonthResponse;
}

export interface MissingDays {
  month: YearMonth;
  days: string[];
}

export interface RatesFetch {
  days: RatesTable;
  missing: MissingDays[];
}

// Fetches every month with a small pool of parallel requests, merging the per-day tables into one. Months with unpublished days are reported separately so the UI can explain them.
export async function fetchRates(
  months: YearMonth[],
  currencies: string[],
  opts: { concurrency?: number; onProgress?: (done: number, total: number) => void; signal?: AbortSignal } = {},
): Promise<RatesFetch> {
  const days: RatesTable = {};
  const missing: MissingDays[] = [];
  const queue = [...months];
  let done = 0;
  const worker = async () => {
    for (let month = queue.shift(); month !== undefined; month = queue.shift()) {
      const res = await fetchRatesMonth(month, currencies, opts.signal);
      Object.assign(days, res.days);
      if (res.missing.length) missing.push({ month, days: res.missing });
      opts.onProgress?.(++done, months.length);
    }
  };
  await Promise.all(Array.from({ length: Math.min(opts.concurrency ?? 6, months.length) }, worker));
  missing.sort((a, b) => a.month.localeCompare(b.month));
  return { days, missing };
}
