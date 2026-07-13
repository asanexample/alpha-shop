// A simple prev/next + numbered pager. Purely presentational — the caller owns what "page" means
// (a server-paginated query param, or a client-side slice of an already-fetched list).
import styles from "./Pager.module.css";

export interface PagerProps {
  page: number; // 1-indexed
  totalPages: number;
  onChange: (page: number) => void;
}

// Windows the page numbers around the current page so a huge catalog doesn't render 100 buttons.
function pageWindow(page: number, totalPages: number): (number | "…")[] {
  const span = 1;
  const pages = new Set<number>([1, totalPages]);
  for (let p = page - span; p <= page + span; p++) {
    if (p >= 1 && p <= totalPages) pages.add(p);
  }
  const sorted = [...pages].sort((a, b) => a - b);
  const out: (number | "…")[] = [];
  for (let i = 0; i < sorted.length; i++) {
    const p = sorted[i] as number;
    const prev = sorted[i - 1];
    if (prev != null && p - prev > 1) out.push("…");
    out.push(p);
  }
  return out;
}

export function Pager({ page, totalPages, onChange }: PagerProps) {
  if (totalPages <= 1) return null;

  return (
    <nav className={styles.pager} aria-label="Pagination">
      <button
        type="button"
        className={styles.nav}
        disabled={page <= 1}
        onClick={() => onChange(page - 1)}
      >
        ← Prev
      </button>

      <div className={styles.pages}>
        {pageWindow(page, totalPages).map((p, i) =>
          p === "…" ? (
            <span key={`e${i}`} className={styles.ellipsis} aria-hidden="true">
              …
            </span>
          ) : (
            <button
              key={p}
              type="button"
              className={styles.page}
              data-active={p === page}
              aria-current={p === page ? "page" : undefined}
              onClick={() => onChange(p)}
            >
              {p}
            </button>
          ),
        )}
      </div>

      <button
        type="button"
        className={styles.nav}
        disabled={page >= totalPages}
        onClick={() => onChange(page + 1)}
      >
        Next →
      </button>
    </nav>
  );
}
