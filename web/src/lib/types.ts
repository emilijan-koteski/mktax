export type Decimal = string;
export type IsoDate = string;
export type IsoTime = string;
export type YearMonth = string;

export type LotStatus = 'taxable' | 'exempt_held_over_2y' | 'out_of_scope_pre_2023' | 'unknown_cost';
export type SkipReason = '' | 'loss' | 'exempt' | 'out_of_scope' | 'unknown_cost';

export interface Warning {
  code: string;
  params: Record<string, string>;
}

export interface Country {
  code: string;
  name_mk: string;
  name_en: string;
}

export interface Lot {
  acquired: IsoDate;
  acquired_at: IsoTime;
  quantity: Decimal;
  cost: Decimal;
  proceeds: Decimal;
  currency: string;
  acquisition_rate: Decimal;
  cost_mkd: Decimal;
  proceeds_mkd: Decimal;
  gain_mkd: Decimal;
  gain_mkd_b: Decimal;
  status: LotStatus;
  tax_free_from: IsoDate;
}

export interface Sale {
  ticker: string;
  isin: string;
  name: string;
  date: IsoDate;
  time: IsoTime;
  ref: string;
  source: string;
  income_type: string;
  income_subtype: string;
  fund_units: boolean;
  country: Country;
  currency: string;
  quantity: Decimal;
  proceeds: Decimal;
  cost: Decimal;
  gain: Decimal;
  taxable_gain: Decimal;
  fees: Decimal;
  sale_rate: Decimal;
  proceeds_mkd: Decimal;
  cost_mkd: Decimal;
  gain_mkd: Decimal;
  gain_mkd_b: Decimal;
  taxable_gain_mkd: Decimal;
  taxable_gain_mkd_b: Decimal;
  exempt_gain_mkd: Decimal;
  out_of_scope_gain_mkd: Decimal;
  tax_mkd: Decimal;
  cost_complete: boolean;
  enter: boolean;
  skip_reason: SkipReason;
  lots: Lot[];
  warnings: Warning[];
}

export interface Dividend {
  ticker: string;
  isin: string;
  name: string;
  type: string;
  date: IsoDate;
  time: IsoTime;
  income_type: string;
  income_subtype: string;
  country: Country;
  gross: Decimal;
  gross_currency: string;
  net: Decimal;
  net_currency: string;
  foreign_tax: Decimal;
  foreign_tax_currency: string;
  gross_rate: Decimal;
  net_rate: Decimal;
  gross_mkd: Decimal;
  net_mkd: Decimal;
  withheld_mkd: Decimal;
  tax_if_credit_accepted_mkd: Decimal;
  tax_if_credit_rejected_mkd: Decimal;
  warnings: Warning[];
}

export interface Totals {
  sales: number;
  unknown_cost_sales: number;
  proceeds_mkd: Decimal;
  taxable_gains_mkd: Decimal;
  taxable_losses_mkd: Decimal;
  net_if_losses_offset_mkd: Decimal;
  tax_base_mkd: Decimal;
  capital_gains_tax_mkd: Decimal;
  tax_if_losses_offset_mkd: Decimal;
  exempt_gains_mkd: Decimal;
  out_of_scope_gains_mkd: Decimal;
  taxable_gains_mkd_b: Decimal;
  capital_gains_tax_mkd_b: Decimal;
  dividends: number;
  dividend_gross_mkd: Decimal;
  dividend_net_mkd: Decimal;
  dividend_withheld_mkd: Decimal;
  dividend_tax_if_credit_accepted_mkd: Decimal;
  dividend_tax_if_credit_rejected_mkd: Decimal;
  tax_due_if_credit_accepted_mkd: Decimal;
  tax_due_if_credit_rejected_mkd: Decimal;
}

export interface PortalRow {
  n: number;
  kind: 'sale' | 'dividend';
  date: IsoDate;
  ticker: string;
  isin: string;
  name: string;
  ref: string;
  income_type: string;
  income_subtype: string;
  country: Country;
  currency: string;
  gross: Decimal;
  foreign_tax: Decimal;
  enter: boolean;
  skip_reason: SkipReason;
  gross_mkd: Decimal;
  tax_mkd: Decimal;
  tax_if_credit_accepted_mkd: Decimal;
}

export interface Month {
  month: YearMonth;
  file_by: IsoDate;
  pay_by: IsoDate;
  has_entries: boolean;
  overdue: boolean;
  in_progress: boolean;
  sales: Sale[];
  dividends: Dividend[];
  totals: Totals;
  portal_rows: PortalRow[];
  warnings: Warning[];
  raw_warnings: string[];
}

export interface OpenLot {
  acquired: IsoDate;
  quantity: Decimal;
  cost: Decimal;
  tax_free_from: IsoDate;
}

export interface OpenPosition {
  key: string;
  ticker: string;
  isin: string;
  name: string;
  currency: string;
  quantity: Decimal;
  cost: Decimal;
  earliest_acquired: IsoDate;
  oldest_lot_tax_free_from: IsoDate;
  tax_free_quantity_now: Decimal;
  lots: OpenLot[];
}

export interface Row {
  file: string;
  line: number;
  id: string;
  action: string;
  date: IsoDate;
  time: IsoTime;
  ticker: string;
  isin: string;
  name: string;
  amount: Decimal;
  currency: string;
  notes: string;
  kind: string;
}

export interface FileSummary {
  name: string;
  rows: number;
  trades: number;
  dividends: number;
  interest: number;
  ignored: number;
  duplicates: number;
  unknown: number;
  first: IsoDate;
  last: IsoDate;
  columns: string[];
}

export interface Instrument {
  key: string;
  ticker: string;
  isin: string;
  name: string;
  currency: string;
  fund_units: boolean;
  fund_units_guessed: boolean;
  buys: number;
  sells: number;
  dividends: number;
}

export interface ImportSummary {
  account_currency: string;
  first: IsoDate;
  last: IsoDate;
  counts: { trades: number; dividends: number };
  files: FileSummary[];
  instruments: Instrument[];
  interest: Row[];
  excluded: Row[];
  ignored: Record<string, number>;
  unknown: Record<string, number>;
  flags: Warning[];
}

export interface Report {
  generated_at: IsoTime;
  currency: string;
  from: YearMonth;
  to: YearMonth;
  fx_source: string;
  fx_method: string;
  months: Month[];
  open_lots: OpenPosition[];
  notes: string[];
  warnings: Warning[];
  raw_warnings: string[];
  import: ImportSummary;
}

export interface RatesNeeded {
  months: YearMonth[];
  currencies: string[];
}

export interface Period {
  from: YearMonth;
  to: YearMonth;
}

export interface ParseResult {
  import: ImportSummary;
  rates_needed: RatesNeeded;
  suggested_period: Period;
}

export interface CalculateResult {
  import: ImportSummary;
  report: Report;
}

export interface MissingRate {
  currency: string;
  date: IsoDate;
}

export type EngineErrorCode = 'bad_input' | 'csv_error' | 'invalid_period' | 'rates_missing' | 'engine_error';

export interface EngineError {
  code: EngineErrorCode;
  message: string;
  missing?: MissingRate[];
}

export type Envelope<T> = { ok: true; result: T } | { ok: false; error: EngineError };

export interface Rate {
  units: number;
  middle: Decimal;
}

export type RatesTable = Record<IsoDate, Record<string, Rate>>;

export interface RatesMonthResponse {
  month: YearMonth;
  from: IsoDate;
  to: IsoDate;
  days: RatesTable;
  missing: IsoDate[];
  complete: boolean;
}

export interface FileIn {
  name: string;
  data: string;
}

export interface CalcOpts {
  from: YearMonth;
  to: YearMonth;
  fund_units: Record<string, boolean>;
}

export interface CalculateInput extends CalcOpts {
  files: FileIn[];
  rates: RatesTable;
  now: string;
}

export interface AppError {
  code: string;
  message: string;
  missing?: MissingRate[];
}
