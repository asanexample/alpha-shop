import { useEffect, useMemo, useState, type FormEvent } from "react";
import { useSearchParams } from "react-router-dom";
import { Pager } from "../components/Pager";
import { ProductGrid } from "../components/ProductGrid";
import { EmptyBlock, ErrorBlock, LoadingBlock } from "../components/States";
import { useProducts } from "../lib/hooks";
import styles from "./Search.module.css";

const PER_PAGE = 24;

function toPage(v: string | null): number {
  const n = v ? parseInt(v, 10) : 1;
  return Number.isFinite(n) && n > 0 ? n : 1;
}

export function Search() {
  const [params, setParams] = useSearchParams();
  const q = params.get("q") ?? "";
  const page = toPage(params.get("page"));
  const [term, setTerm] = useState(q);
  useEffect(() => setTerm(q), [q]);

  // With no query we list everything (a plain "browse all") — the biggest listing in the shop, so
  // this is where server-side pagination matters most.
  const query = useMemo(
    () => ({ ...(q ? { q } : {}), page, perPage: PER_PAGE }),
    [q, page],
  );
  const { data, isLoading, isError, error, refetch, isFetching } = useProducts(query);
  const products = data?.products ?? [];
  const totalPages = data ? Math.max(1, Math.ceil(data.total / data.perPage)) : 1;

  function submit(e: FormEvent) {
    e.preventDefault();
    const next = new URLSearchParams(params);
    const v = term.trim();
    if (v) next.set("q", v);
    else next.delete("q");
    next.delete("page"); // a new search always starts back at page 1
    setParams(next, { replace: true });
  }

  function goToPage(p: number) {
    const next = new URLSearchParams(params);
    if (p > 1) next.set("page", String(p));
    else next.delete("page");
    setParams(next, { replace: true });
  }

  return (
    <div className={styles.wrap}>
      <div className={styles.head}>
        <p className="eyebrow">Search</p>
        <h1 className={styles.title}>
          {q ? <>Results for “{q}”</> : "Browse the shop"}
        </h1>
        <form role="search" onSubmit={submit} className={styles.form}>
          <input
            type="search"
            name="q"
            aria-label="Search products"
            placeholder="Try “gravel”, “helmet”, or a brand"
            value={term}
            onChange={(e) => setTerm(e.target.value)}
          />
          <button type="submit" className="btn">
            Search
          </button>
        </form>
        <span className={styles.count} aria-live="polite">
          {isLoading ? "Searching…" : `${data?.total ?? 0} ${data?.total === 1 ? "result" : "results"}`}
          {isFetching && !isLoading ? " · updating" : ""}
        </span>
      </div>

      {isLoading ? (
        <LoadingBlock label="Searching…" />
      ) : isError ? (
        <ErrorBlock error={error} onRetry={() => refetch()} />
      ) : products.length === 0 ? (
        <EmptyBlock title="No matches">
          {q ? (
            <>Nothing matched “{q}”. Check the spelling, or try a broader term like a category or brand.</>
          ) : (
            <>Type something above to search the catalog.</>
          )}
        </EmptyBlock>
      ) : (
        <>
          <ProductGrid products={products} />
          <Pager page={page} totalPages={totalPages} onChange={goToPage} />
        </>
      )}
    </div>
  );
}
