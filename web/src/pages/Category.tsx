import { useEffect, useMemo, useState, type FormEvent } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { Breadcrumb } from "../components/Breadcrumb";
import { ProductGrid } from "../components/ProductGrid";
import { EmptyBlock, ErrorBlock, LoadingBlock } from "../components/States";
import { useNav, useNavLookups, useProducts } from "../lib/hooks";
import { KIND_LABEL, type Product } from "../lib/types";
import styles from "./Category.module.css";

interface Bracket {
  label: string;
  min?: number;
  max?: number;
}
const BRACKETS: Bracket[] = [
  { label: "Under $100", max: 100 },
  { label: "$100–500", min: 100, max: 500 },
  { label: "$500–2,000", min: 500, max: 2000 },
  { label: "$2,000+", min: 2000 },
];

function toInt(v: string | null): number | undefined {
  if (!v) return undefined;
  const n = parseInt(v, 10);
  return Number.isFinite(n) && n > 0 ? n : undefined;
}

export function Category() {
  const { slug } = useParams<{ slug: string }>();
  const [params, setParams] = useSearchParams();
  const { data: nav } = useNav();
  const { category } = useNavLookups();
  const cat = slug ? category(slug) : undefined;

  const [showFilters, setShowFilters] = useState(false);

  // ---- filter state derived from the URL ----
  const selectedBrands = useMemo(() => new Set(params.getAll("brand")), [params]);
  const minDollar = params.get("min");
  const maxDollar = params.get("max");
  const onSale = params.get("sale") === "1";

  const query = useMemo(
    () => ({
      category: slug,
      onSale,
      minPrice: toInt(minDollar),
      maxPrice: toInt(maxDollar),
    }),
    [slug, onSale, minDollar, maxDollar],
  );

  const { data, isLoading, isError, error, refetch, isFetching } = useProducts(query);

  // Brand is multi-select (checkboxes); the catalog API filters a single brand, so we apply the
  // brand set client-side over the price/sale-filtered results the API returns.
  const apiProducts = data?.products ?? [];
  const products = useMemo(
    () => (selectedBrands.size ? apiProducts.filter((p) => selectedBrands.has(p.brand)) : apiProducts),
    [apiProducts, selectedBrands],
  );

  // Brand facet counts within the current price/sale filter.
  const brandCounts = useMemo(() => {
    const m = new Map<string, number>();
    for (const p of apiProducts) m.set(p.brand, (m.get(p.brand) ?? 0) + 1);
    return m;
  }, [apiProducts]);

  const brandsToShow = useMemo(
    () => (nav?.brands ?? []).filter((b) => (brandCounts.get(b.slug) ?? 0) > 0 || selectedBrands.has(b.slug)),
    [nav?.brands, brandCounts, selectedBrands],
  );

  // ---- mutators ----
  function commit(mutate: (p: URLSearchParams) => void) {
    const next = new URLSearchParams(params);
    mutate(next);
    setParams(next, { replace: true });
  }

  function toggleBrand(brandSlug: string) {
    commit((p) => {
      const set = new Set(p.getAll("brand"));
      p.delete("brand");
      if (set.has(brandSlug)) set.delete(brandSlug);
      else set.add(brandSlug);
      for (const b of set) p.append("brand", b);
    });
  }

  function setBracket(b: Bracket) {
    commit((p) => {
      const active = String(minDollar ?? "") === String(b.min ?? "") && String(maxDollar ?? "") === String(b.max ?? "");
      p.delete("min");
      p.delete("max");
      if (!active) {
        if (b.min) p.set("min", String(b.min));
        if (b.max) p.set("max", String(b.max));
      }
    });
  }

  function clearAll() {
    setParams(new URLSearchParams(), { replace: true });
  }

  const hasFilters = selectedBrands.size > 0 || !!minDollar || !!maxDollar || onSale;

  // ---- price inputs (local while typing, committed on blur/submit) ----
  const [minLocal, setMinLocal] = useState(minDollar ?? "");
  const [maxLocal, setMaxLocal] = useState(maxDollar ?? "");
  useEffect(() => setMinLocal(minDollar ?? ""), [minDollar]);
  useEffect(() => setMaxLocal(maxDollar ?? ""), [maxDollar]);

  function commitPrice(e?: FormEvent) {
    e?.preventDefault();
    commit((p) => {
      const mn = parseInt(minLocal, 10);
      const mx = parseInt(maxLocal, 10);
      if (Number.isFinite(mn) && mn > 0) p.set("min", String(mn));
      else p.delete("min");
      if (Number.isFinite(mx) && mx > 0) p.set("max", String(mx));
      else p.delete("max");
    });
  }

  const title = cat?.name ?? slug ?? "Shop";
  const kindLabel = cat ? KIND_LABEL[cat.kind] : "Shop";

  return (
    <div className={styles.wrap}>
      <Breadcrumb
        items={[{ label: "Home", to: "/" }, { label: kindLabel }, { label: title }]}
      />

      <header className={styles.head}>
        <div>
          <h1 className={styles.title}>{title}</h1>
          {cat?.blurb ? <p className={styles.blurb}>{cat.blurb}</p> : null}
        </div>
      </header>

      <button
        type="button"
        className={styles.filterToggle}
        aria-expanded={showFilters}
        onClick={() => setShowFilters((v) => !v)}
      >
        {showFilters ? "Hide filters" : "Filters"}
        {hasFilters ? ` · ${selectedBrands.size + (minDollar || maxDollar ? 1 : 0) + (onSale ? 1 : 0)}` : ""}
      </button>

      <div className={styles.body}>
        {/* ---- FACET SIDEBAR ---- */}
        <aside className={styles.sidebar} data-open={showFilters} aria-label="Filters">
          <div className={styles.facet}>
            <div className={styles.facetTitle}>Brand</div>
            {brandsToShow.length === 0 ? (
              <p className="mono" style={{ fontSize: "0.8rem", color: "var(--muted)" }}>
                No brands to filter.
              </p>
            ) : (
              brandsToShow.map((b) => (
                <label key={b.slug} className={styles.check}>
                  <input
                    type="checkbox"
                    checked={selectedBrands.has(b.slug)}
                    onChange={() => toggleBrand(b.slug)}
                  />
                  <span className={styles.cName}>{b.name}</span>
                  <span className={styles.cCount}>{brandCounts.get(b.slug) ?? 0}</span>
                </label>
              ))
            )}
          </div>

          <div className={styles.facet}>
            <div className={styles.facetTitle}>Price (USD)</div>
            <form className={styles.price} onSubmit={commitPrice}>
              <span className={styles.priceField}>
                <span>$</span>
                <input
                  type="number"
                  inputMode="numeric"
                  min={0}
                  placeholder="Min"
                  aria-label="Minimum price in dollars"
                  value={minLocal}
                  onChange={(e) => setMinLocal(e.target.value)}
                  onBlur={() => commitPrice()}
                />
              </span>
              <span className={styles.priceDash} aria-hidden="true">
                –
              </span>
              <span className={styles.priceField}>
                <span>$</span>
                <input
                  type="number"
                  inputMode="numeric"
                  min={0}
                  placeholder="Max"
                  aria-label="Maximum price in dollars"
                  value={maxLocal}
                  onChange={(e) => setMaxLocal(e.target.value)}
                  onBlur={() => commitPrice()}
                />
              </span>
              <button type="submit" className="sr-only">
                Apply price
              </button>
            </form>
            <div className={styles.brackets}>
              {BRACKETS.map((b) => {
                const active =
                  String(minDollar ?? "") === String(b.min ?? "") &&
                  String(maxDollar ?? "") === String(b.max ?? "");
                return (
                  <button
                    key={b.label}
                    type="button"
                    className={styles.bracket}
                    data-active={active}
                    onClick={() => setBracket(b)}
                  >
                    {b.label}
                  </button>
                );
              })}
            </div>
          </div>

          <div className={styles.facet}>
            <div className={styles.facetTitle}>Offers</div>
            <label className={styles.toggle}>
              <input
                type="checkbox"
                checked={onSale}
                onChange={() =>
                  commit((p) => (onSale ? p.delete("sale") : p.set("sale", "1")))
                }
              />
              <span>On sale only</span>
            </label>
          </div>

          {hasFilters ? (
            <button type="button" className={styles.clear} onClick={clearAll}>
              Clear all filters
            </button>
          ) : null}
        </aside>

        {/* ---- RESULTS ---- */}
        <section className={styles.results} aria-label="Products">
          <div className={styles.resultsTop}>
            <div className={styles.activeChips}>
              {[...selectedBrands].map((bs) => {
                const name = nav?.brands.find((b) => b.slug === bs)?.name ?? bs;
                return (
                  <span key={bs} className={styles.chip}>
                    {name}
                    <button type="button" aria-label={`Remove ${name}`} onClick={() => toggleBrand(bs)}>
                      ×
                    </button>
                  </span>
                );
              })}
              {onSale ? (
                <span className={styles.chip}>
                  On sale
                  <button type="button" aria-label="Remove on sale filter" onClick={() => commit((p) => p.delete("sale"))}>
                    ×
                  </button>
                </span>
              ) : null}
              {minDollar || maxDollar ? (
                <span className={styles.chip}>
                  {minDollar ? `$${minDollar}` : "$0"}–{maxDollar ? `$${maxDollar}` : "∞"}
                  <button
                    type="button"
                    aria-label="Remove price filter"
                    onClick={() => commit((p) => {
                      p.delete("min");
                      p.delete("max");
                    })}
                  >
                    ×
                  </button>
                </span>
              ) : null}
            </div>
            <span className={styles.count} aria-live="polite">
              {isLoading ? "…" : `${products.length} ${products.length === 1 ? "result" : "results"}`}
              {isFetching && !isLoading ? " · updating" : ""}
            </span>
          </div>

          {isLoading ? (
            <LoadingBlock />
          ) : isError ? (
            <ErrorBlock error={error} onRetry={() => refetch()} />
          ) : products.length === 0 ? (
            <EmptyBlock title="No matches">
              {hasFilters ? (
                <>
                  Nothing here fits those filters. Try widening the price range or clearing a brand —{" "}
                  <button
                    type="button"
                    onClick={clearAll}
                    style={{
                      background: "none",
                      border: "none",
                      color: "var(--ink)",
                      textDecoration: "underline",
                      cursor: "pointer",
                      font: "inherit",
                      padding: 0,
                    }}
                  >
                    clear all filters
                  </button>
                  .
                </>
              ) : (
                <>We're between shipments in this category. Check back soon or browse another section.</>
              )}
            </EmptyBlock>
          ) : (
            <ProductGrid products={products as Product[]} />
          )}
        </section>
      </div>
    </div>
  );
}
