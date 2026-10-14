const nf2 = new Intl.NumberFormat('mk-MK', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
const nfQty = new Intl.NumberFormat('mk-MK', { minimumFractionDigits: 0, maximumFractionDigits: 8 });
const monthName = new Intl.DateTimeFormat('mk-MK', { month: 'long', timeZone: 'UTC' });

export const collator = new Intl.Collator('mk');

export const fmtMkd = (d: string): string => `${nf2.format(Number(d))}\u00a0ден.`;

export const fmtAmount = (d: string, currency: string): string => `${nf2.format(Number(d))}\u00a0${currency}`;

export const fmtQty = (d: string): string => nfQty.format(Number(d));

export const fmtNumber = (d: string): string => nf2.format(Number(d));

export function fmtDate(iso: string): string {
  if (!iso) return '';
  const [y, m, d] = iso.split('-');
  return `${d}.${m}.${y}`;
}

export function fmtMonth(ym: string): string {
  if (!ym) return '';
  const [y, m] = ym.split('-').map(Number);
  return `${monthName.format(new Date(Date.UTC(y, m - 1, 1)))} ${y}`;
}

export function fmtDateTime(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${pad(d.getDate())}.${pad(d.getMonth() + 1)}.${d.getFullYear()} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export const isNegative = (d: string): boolean => d.trim().startsWith('-');

export const isZero = (d: string): boolean => /^-?0(\.0+)?$/.test(d.trim());
