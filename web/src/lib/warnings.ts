import { m, t } from './i18n';
import type { AppError, Warning } from './types';

export function warningText(w: Warning): string {
  const tpl = (m.warnings as Record<string, string>)[w.code];
  return tpl ? t(tpl, w.params) : `${w.code}: ${Object.values(w.params).join(', ')}`;
}

export function errorText(e: AppError): string {
  const tpl = (m.errors as Record<string, string>)[e.code];
  const first = e.missing?.[0];
  const params = { message: e.message, n: e.missing?.length ?? 0, first: first ? `${first.currency} ${first.date}` : '' };
  return tpl ? t(tpl, params) : t(m.errors.unknown, { message: `${e.code}: ${e.message}` });
}
