import { useEffect, useState, type FormEvent } from "react";
import { useSearchParams } from "react-router-dom";
import { ProductGrid } from "../components/ProductGrid";
import { EmptyBlock, ErrorBlock, LoadingBlock } from "../components/States";
import { useProducts } from "../lib/hooks";
import styles from "./Search.module.css";

export function Search() {
  const [params, setParams] = useSearchParams();
  const q = params.get("q") ?? "";
  const [term, setTerm] = useState(q);
  useEffect(() => setTerm(q), [q]);

  // With no query we list everything (a plain "browse all").
  const { data, isLoading, isError, error, refetch } = useProducts(q ? { q } : {});
  const products = data?.products ?? [];

  function submit(e: FormEvent) {
    e.preventDefault();
    const next = new URLSearchParams(params);
    const v = term.trim();
    if (v) next.set("q", v);
    else next.delete("q");
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
          {isLoading ? "Searching…" : `${products.length} ${products.length === 1 ? "result" : "results"}`}
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
        <ProductGrid products={products} />
      )}
    </div>
  );
}
