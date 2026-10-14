import { m } from '../i18n';
import type { Report } from '../types';

const decimal = (s: string) => s.replace('.', ',');

function quote(v: string | number): string {
  const s = String(v);
  return /[;"\r\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s;
}

export function portalRowsCsv(report: Report): string {
  const h = m.csv.headers;
  const header = [h.month, h.fileBy, h.payBy, h.n, h.kind, h.ticker, h.isin, h.name, h.date, h.incomeType, h.incomeSubtype, h.countryCode, h.country, h.currency, h.gross, h.foreignTax, h.enter, h.skipReason, h.grossMkd, h.taxMkd, h.ref];
  const lines: (string | number)[][] = [header];
  for (const month of report.months) {
    for (const r of month.portal_rows) {
      lines.push([
        month.month, month.file_by, month.pay_by, r.n, r.kind === 'sale' ? m.portal.sale : m.portal.dividend,
        r.ticker, r.isin, r.name, r.date, r.income_type, r.income_subtype, r.country.code, r.country.name_mk, r.currency,
        decimal(r.gross), decimal(r.foreign_tax), r.enter ? m.portal.yes : m.portal.no, r.skip_reason ? m.portal.skip[r.skip_reason] : '',
        decimal(r.gross_mkd), decimal(r.tax_mkd), r.ref,
      ]);
    }
  }
  return '﻿' + lines.map((l) => l.map(quote).join(';')).join('\r\n') + '\r\n';
}

export function downloadText(name: string, text: string, type = 'text/csv;charset=utf-8'): void {
  const url = URL.createObjectURL(new Blob([text], { type }));
  const a = document.createElement('a');
  a.href = url;
  a.download = name;
  document.body.append(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}
