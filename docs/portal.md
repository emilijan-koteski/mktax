# Entering the figures in the UJP portal

This is how the public e-ПДД portal looked on 14 October 2026. Check the portal itself if anything differs; it is the authority.

1. Sign in at `etax.ujp.gov.mk` and open the **ПДД** application.
2. In the left menu choose **Аконтативна пресметка → Е-ПДД → Внес**. Do not use **Капитални добивки → Внес**: that form covers domestic property and shares in companies, not foreign securities.
3. Fill in the submitter header. **Од датум / До датум** is the income period, normally one calendar month. **Начин на исплата** is **Друг начин**.
4. Add one item (ставка) per row of the report's portal table:
   - **Држава во која е резидент исплатувач**: the payer's country. Choosing a country other than Macedonia opens the foreign-currency fields.
   - **Вид доход / Подвид доход**: `T8 Капитални добивки` with `S8.1` for shares or `S.8.7` for fund units (ETFs); `T7 Доход од капитал` with `S7.1` for dividends.
   - **Валута** and **Бруто валута**: for a sale, the gain in the sale currency; for a dividend, the gross dividend.
   - **Платен данок валута**: the foreign tax withheld on a dividend; zero for sales.
   - **Датум на плаќање**: the transaction date.
5. The portal applies the 10% normed costs and converts to denars itself, so enter the gain, not the tax.
6. Attach documents (the printed report and the broker export) to each item, save, close the calculation and sign it with the PIN sent by e-mail.

Deadlines: file by the 10th and pay by the 15th of the month after the income. Late filing is possible by setting the period to a past month.

Open questions the app lists under "Забелешки и толкувања" (conversion method, pre-2023 lots, foreign losses, the dividend credit proof, normed costs on foreign securities, the dividend subtype, reporting unit, fees, corporate actions, accumulating ETFs, penalties) are not settled by UJP guidance as of the date above.
