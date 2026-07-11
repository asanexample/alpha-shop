import { useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useCart } from "../context/CartContext";
import { useNav } from "../lib/hooks";
import { KIND_LABEL, KIND_ORDER, type Category, type Kind } from "../lib/types";
import { CartIcon, ChevronDown, Chainring, CloseIcon, MenuIcon, SearchIcon } from "./Glyphs";
import styles from "./Header.module.css";

function useGroupedCategories(categories: Category[] | undefined) {
  return useMemo(() => {
    const byKind = new Map<Kind, Category[]>();
    for (const c of categories ?? []) {
      const list = byKind.get(c.kind) ?? [];
      list.push(c);
      byKind.set(c.kind, list);
    }
    // Keep the fixed nav order; drop kinds with no categories.
    return KIND_ORDER.filter((k) => byKind.has(k)).map((k) => ({
      kind: k,
      label: KIND_LABEL[k],
      categories: byKind.get(k) as Category[],
    }));
  }, [categories]);
}

export function Header() {
  const { data } = useNav();
  const groups = useGroupedCategories(data?.categories);
  const { count, notify } = useCart();
  const navigate = useNavigate();

  const [openKind, setOpenKind] = useState<Kind | null>(null);
  const [drawer, setDrawer] = useState(false);
  const [term, setTerm] = useState("");
  const navRef = useRef<HTMLElement>(null);

  // Hover intent: opening is immediate, but closing is delayed so the pointer can cross the small gap
  // between a top item and its dropdown (and move diagonally onto a sub-item) without the menu snapping
  // shut. Re-entering any nav item cancels the pending close.
  const closeTimer = useRef<number | null>(null);
  function cancelClose() {
    if (closeTimer.current !== null) {
      window.clearTimeout(closeTimer.current);
      closeTimer.current = null;
    }
  }
  function openMenu(kind: Kind) {
    cancelClose();
    setOpenKind(kind);
  }
  function scheduleClose() {
    cancelClose();
    closeTimer.current = window.setTimeout(() => setOpenKind(null), 220);
  }
  useEffect(() => cancelClose, []);

  // Escape closes any open menu / the drawer.
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") {
        setOpenKind(null);
        setDrawer(false);
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  // Lock body scroll while the drawer is open.
  useEffect(() => {
    document.body.style.overflow = drawer ? "hidden" : "";
    return () => {
      document.body.style.overflow = "";
    };
  }, [drawer]);

  function submitSearch(e: FormEvent) {
    e.preventDefault();
    const q = term.trim();
    if (!q) return;
    setDrawer(false);
    navigate(`/search?q=${encodeURIComponent(q)}`);
    setTerm("");
  }

  return (
    <header className={styles.header}>
      {/* thin ink utility bar */}
      <div className={styles.utility}>
        <div className={styles.utilityInner}>
          <span>
            Alpha Bikes <span className={styles.utilityNote}>— Portland, OR</span>
          </span>
          <span className={styles.utilityNote}>
            Free shipping over <strong>$99</strong> · In-store pickup daily
          </span>
        </div>
      </div>

      {/* main bar */}
      <div className={styles.bar}>
        <Link to="/" className={styles.wordmark} aria-label="Alpha Bikes — home">
          <Chainring className={styles.ring} />
          <span>
            <span className={styles.alpha}>Alpha</span> Bikes
          </span>
        </Link>

        <nav
          className={styles.nav}
          aria-label="Shop categories"
          ref={navRef}
          onBlur={(e) => {
            // Close when focus leaves the whole nav (keyboard).
            if (!navRef.current?.contains(e.relatedTarget as Node)) setOpenKind(null);
          }}
        >
          {groups.map((g) => {
            const isOpen = openKind === g.kind;
            return (
              <div
                key={g.kind}
                className={styles.navItem}
                data-open={isOpen}
                onMouseEnter={() => openMenu(g.kind)}
                onMouseLeave={scheduleClose}
                onFocus={() => openMenu(g.kind)}
              >
                <button
                  type="button"
                  className={styles.navTop}
                  aria-expanded={isOpen}
                  aria-haspopup="true"
                  onClick={() => setOpenKind(isOpen ? null : g.kind)}
                >
                  {g.label}
                  <ChevronDown className={styles.chev} />
                </button>
                {isOpen ? (
                  <div className={styles.panel} role="menu" aria-label={g.label}>
                    <div className={styles.panelHead}>{g.label}</div>
                    {g.categories.map((c) => (
                      <Link
                        key={c.slug}
                        to={`/c/${c.slug}`}
                        className={styles.panelLink}
                        role="menuitem"
                        onClick={() => setOpenKind(null)}
                      >
                        <span className={styles.panelLinkName}>{c.name}</span>
                        {c.blurb ? <span className={styles.panelLinkBlurb}>{c.blurb}</span> : null}
                      </Link>
                    ))}
                  </div>
                ) : null}
              </div>
            );
          })}
        </nav>

        <div className={styles.actions}>
          <form className={styles.search} role="search" onSubmit={submitSearch}>
            <SearchIcon />
            <input
              type="search"
              name="q"
              placeholder="Search the shop"
              aria-label="Search products"
              value={term}
              onChange={(e) => setTerm(e.target.value)}
            />
          </form>

          <button
            type="button"
            className={styles.iconBtn}
            aria-label={`Cart — ${count} item${count === 1 ? "" : "s"}`}
            onClick={() => notify("Your cart is a preview — checkout is coming soon.")}
          >
            <CartIcon />
            {count > 0 ? <span className={styles.cartCount}>{count}</span> : null}
          </button>

          <button
            type="button"
            className={`${styles.iconBtn} ${styles.menuToggle}`}
            aria-label="Open menu"
            aria-expanded={drawer}
            onClick={() => setDrawer(true)}
          >
            <MenuIcon />
          </button>
        </div>
      </div>

      {/* mobile drawer */}
      {drawer ? (
        <>
          <div className={styles.backdrop} onClick={() => setDrawer(false)} aria-hidden="true" />
          <div className={styles.drawer} role="dialog" aria-modal="true" aria-label="Menu">
            <div className={styles.drawerHead}>
              <span className={styles.wordmark}>
                <Chainring className={styles.ring} /> Alpha Bikes
              </span>
              <button
                type="button"
                className={styles.iconBtn}
                aria-label="Close menu"
                onClick={() => setDrawer(false)}
              >
                <CloseIcon />
              </button>
            </div>
            <div className={styles.drawerBody}>
              <form className={styles.drawerSearch} role="search" onSubmit={submitSearch}>
                <SearchIcon />
                <input
                  type="search"
                  name="q"
                  placeholder="Search the shop"
                  aria-label="Search products"
                  value={term}
                  onChange={(e) => setTerm(e.target.value)}
                />
              </form>
              {groups.map((g) => (
                <div key={g.kind} className={styles.drawerGroup}>
                  <div className={styles.drawerGroupTitle}>{g.label}</div>
                  {g.categories.map((c) => (
                    <Link
                      key={c.slug}
                      to={`/c/${c.slug}`}
                      className={styles.drawerLink}
                      onClick={() => setDrawer(false)}
                    >
                      {c.name}
                    </Link>
                  ))}
                </div>
              ))}
            </div>
          </div>
        </>
      ) : null}
    </header>
  );
}
