import { m } from './i18n';

export const collator = new Intl.Collator('mk');

function group(int: string): string {
  return int.replace(/\B(?=(\d{3})+(?!\d))/g, '.');
}

// Formats a decimal string in the Macedonian convention (thousands dot, decimal comma) without relying on the browser's ICU data for the mk locale.
export function fmtNumber(d: string, minFraction = 2, maxFraction = 2): string {
  const s = d.trim();
  if (s === '' || !/^-?\d+(\.\d+)?$/.test(s)) return s;
  const negative = s.startsWith('-');
  const [int, frac = ''] = s.replace('-', '').split('.');
  let fraction = frac.slice(0, maxFraction).replace(/0+$/, '');
  if (maxFraction === minFraction) fraction = frac.slice(0, maxFraction).padEnd(maxFraction, '0');
  else if (fraction.length < minFraction) fraction = fraction.padEnd(minFraction, '0');
  const body = fraction ? `${group(int)},${fraction}` : group(int);
  return negative && !/^[0.,]*$/.test(body) ? `-${body}` : body;
}

export const fmtMkd = (d: string): string => `${fmtNumber(d)}\u00a0ден.`;

export const fmtAmount = (d: string, currency: string): string => `${fmtNumber(d)}\u00a0${currency}`;

export const fmtQty = (d: string): string => fmtNumber(d, 0, 8);

export function fmtDate(iso: string): string {
  if (!iso) return '';
  const [y, mo, d] = iso.split('-');
  return `${d}.${mo}.${y}`;
}

export function fmtMonth(ym: string): string {
  if (!ym) return '';
  const [y, mo] = ym.split('-').map(Number);
  return `${m.months[mo - 1]} ${y}`;
}

export function fmtDateTime(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${pad(d.getDate())}.${pad(d.getMonth() + 1)}.${d.getFullYear()} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export const isNegative = (d: string): boolean => d.trim().startsWith('-');

export const isZero = (d: string): boolean => /^-?0(\.0+)?$/.test(d.trim());

export function sumDecimals(values: string[]): string {
  const scale = Math.max(2, ...values.map((v) => v.trim().split('.')[1]?.length ?? 0));
  let total = 0n;
  for (const v of values) {
    const s = v.trim();
    if (!/^-?\d+(\.\d+)?$/.test(s)) continue;
    const negative = s.startsWith('-');
    const [int, frac = ''] = s.replace('-', '').split('.');
    const n = BigInt(int + frac.padEnd(scale, '0'));
    total += negative ? -n : n;
  }
  const negative = total < 0n;
  const digits = (negative ? -total : total).toString().padStart(scale + 1, '0');
  return `${negative ? '-' : ''}${digits.slice(0, -scale)}.${digits.slice(-scale)}`;
}
