# Tax rules encoded in `mktax/`

The engine computes North Macedonia personal income tax figures on foreign securities for the monthly e-ППД self-assessment: capital gains (FIFO lots, two-year exemption, 10% normed costs) and foreign dividends (10% of gross, with and without the foreign tax credit), converted to MKD at the National Bank (NBRSM) middle rate.

The package is pure: no I/O and no clock (the caller passes `Options.Now`). It imports only the standard library and `shopspring/decimal`. Callers adapt their broker data into `Trade` and `Dividend` values and supply a `RateFunc`.

ЗДЛД = Закон за данокот на личен доход.

## Rules

- Gain = sale price − purchase price, on sale or exchange (Art. 58), for securities acquired on or after 2023-01-01 (transitional Art. 21 of Sl. vesnik 274/22). Older lots are consumed by FIFO but their gain is treated as out of scope. This is an interpretation; confirm it with UJP.
- FIFO per instrument, oldest lot first (Art. 22 of Sl. vesnik 274/22). Fractional quantities are supported; a partially used lot keeps its date and the rest of its cost, and proceeds are split pro rata with the remainder on the last portion so the parts add up exactly.
- Buys are matched before sells at the same instant. A missing ISIN is filled from another trade of the same ticker so one instrument never splits into two FIFO queues. The FIFO key is the ISIN, else the ticker.
- A lot portion is exempt when held longer than two years (Art. 12(1) item 39-а): the sale date must be strictly after the acquisition date plus two calendar years, tested per lot on Europe/Skopje dates. `AddDate` semantics apply, so 29 February plus two years is 1 March and the lot is exempt from 2 March.
- Tax base is 90% of the gain (10% normed costs, Art. 60(2)); the rate is 10% (Art. 11), so 9% effective. Fees are informational and never added to the cost, because the normed costs cover them.
- Foreign losses are not netted: Art. 62 allows it only for sales on a domestic exchange. `NetIfLossesOffsetMKD` and `TaxIfLossesOffsetMKD` are information only.
- Dividends are taxed at 10% of the gross amount (Art. 55–57). The foreign tax credit (Art. 8(4)) is capped at 10% and shown both accepted and rejected, because UJP currently accepts it only with proof from the foreign tax authority. When the caller does not know the gross, the net is used as gross and a warning is added.
- Withheld tax is derived as max(0, GrossMKD − NetMKD), with gross and net each converted at the payment-date rate in its own currency.
- Dates (periods, holding periods, deadlines) are Europe/Skopje calendar dates. `time/tzdata` is embedded so the zone resolves on any host, including WebAssembly.
- Deadlines: file the e-ППД by the 10th and pay by the 15th of the month after the income. They are not shifted for weekends or public holidays. A month is overdue when it has entries and today is after the filing date.

## MKD conversion

Rates are NBRSM middle rates valid on each transaction's date. Tax figures use method A (`per_leg`): proceeds at the sale-date rate and each lot's cost at its acquisition-date rate. Method B (`gain_at_sale_rate`) converts (proceeds − cost) at the sale-date rate and is shown for comparison only; when a lot was bought in a different currency than the sale, method B is undefined and repeats method A with a warning. UJP has not specified the method.

Sales outside the requested months still consume lots, and the rate for every matched lot's acquisition date is required, including lots acquired before 2023.

## e-ППД codes

Spelled as the portal lists them (Latin T/S), read from the portal's own calculator on 2026-10-14:

| Income | Вид доход | Подвид доход |
| --- | --- | --- |
| Share sale | T8 Капитални добивки | S8.1 Капитална добивка од продажба на хартии од вредност |
| Fund-unit (ETF) sale | T8 Капитални добивки | S.8.7 Капитални добивки од продажба на удели издадени од инвестициски фондови |
| Dividend | T7 Доход од капитал | S7.1 Дивиденди и друг доход остварени со учество во добивката |

Choose the payer's country (not Macedonia) to open the foreign-currency fields. Enter the gain itself: the portal applies the 10% normed costs (1,000 gain → 90 tax). The payer country comes from the ISIN prefix, which can differ from the issuer's tax residence (ADRs carry a US ISIN).

## Field semantics

- `Trade.Value` is positive and in `Trade.Currency`: the cost for a buy, the proceeds for a sell, as the amount that moved in the account. `Trade.Source` and `Trade.Ref` are echoed back. `Trade.FundUnits` marks investment-fund units (ETF/UCITS), reported under S.8.7 instead of S8.1.
- `Dividend.Gross` is the amount before withholding, zero if unknown; `Dividend.Net` is the amount received. `Dividend.Type` is a caller label.
- `RateFunc` returns MKD per one unit of a currency on a day and is never called for MKD.
- `LotStatus`: `taxable`, `exempt_held_over_2y`, `out_of_scope_pre_2023`, `unknown_cost` (sold more than the recorded buys).
- `Sale.Cost`, `Sale.Gain` and `Sale.TaxableGain` are in the account currency without MKD conversion and cover matched lots only. `Sale.ProceedsMKD` covers the whole sale, including any unmatched part. `Sale.TaxMKD` = max(0, TaxableGainMKD) × 90% × 10%.
- `MonthTotals.TaxableGainsMKD` sums the positive per-sale taxable gains; `TaxableLossesMKD` the negative ones (not deductible). Capital gains tax uses method A; the B figures are for comparison.
- `OpenPosition.TaxFreeQuantityNow` is the quantity in lots already held over two years at `Now`; lots are oldest first. Open positions are sorted by ticker, then key.
- `Report.Warnings` lists sales exceeding the recorded buys anywhere in history; such a sale is reported with an unknown-cost portion rather than failing the report.
- Quantities at or below 1e-8 are treated as rounding dust.

## Tests

All figures in the test files are synthetic. Tickers are placeholders and ISINs are a country prefix followed by zeros.
