import { mk } from './mk';

export type Messages = typeof mk;

export const m: Messages = mk;

const rules = new Intl.PluralRules('mk');

export function t(tpl: string, params: Record<string, string | number> = {}): string {
  return tpl.replace(/\{(\w+)\}/g, (_, k: string) => (k in params ? String(params[k]) : `{${k}}`));
}

export function plural(n: number, forms: { one: string; other: string }): string {
  return t(rules.select(n) === 'one' ? forms.one : forms.other, { n });
}

export function keyPaths(obj: unknown, prefix = ''): string[] {
  if (typeof obj !== 'object' || obj === null) return [prefix];
  return Object.entries(obj as Record<string, unknown>).flatMap(([k, v]) => keyPaths(v, prefix ? `${prefix}.${k}` : k));
}
