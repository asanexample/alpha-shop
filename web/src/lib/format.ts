// Price/currency formatting. Prices are whole USD cents on the wire.

const withCents = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

const whole = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
  minimumFractionDigits: 0,
  maximumFractionDigits: 0,
});

/**
 * Format cents as USD. Whole-dollar amounts drop the decimals (420000 → "$4,200"),
 * amounts with cents keep them (8999 → "$89.99").
 */
export function formatCents(cents: number): string {
  const dollars = cents / 100;
  return Number.isInteger(dollars) ? whole.format(dollars) : withCents.format(dollars);
}

/** Whole dollars (rounded) — used for the facet price inputs. */
export function centsToDollars(cents: number): number {
  return Math.round(cents / 100);
}

/** Percent saved when on sale, e.g. 12 → "-12%". */
export function savingsLabel(priceCents: number, saleCents: number): string {
  const pct = Math.round((1 - saleCents / priceCents) * 100);
  return `-${pct}%`;
}

const dateFmt = new Intl.DateTimeFormat("en-US", { dateStyle: "medium" });

/** Format an ISO timestamp (e.g. an order's createdAt) as a short date, e.g. "Jul 12, 2026". */
export function formatDate(iso: string): string {
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? iso : dateFmt.format(d);
}
