import { useMemo } from "react";
import { Link } from "react-router-dom";
import { Breadcrumb } from "../components/Breadcrumb";
import { ErrorBlock, LoadingBlock } from "../components/States";
import { Thumb } from "../components/Thumb";
import { CloseIcon } from "../components/Glyphs";
import { useCart } from "../context/CartContext";
import { formatCents } from "../lib/format";
import { useNav } from "../lib/hooks";
import type { Brand, CartLine } from "../lib/types";
import styles from "./Cart.module.css";

const FREE_SHIPPING_CENTS = 9900;

// The cart line carries no brand slug, so resolve a brand-name eyebrow from nav when the product name
// begins with a known brand (e.g. "Trek Domane SL 6" → "Trek"). Graceful: returns null when unresolved.
function makeBrandResolver(brands: Brand[] | undefined) {
  const sorted = [...(brands ?? [])].sort((a, b) => b.name.length - a.name.length);
  return (line: CartLine): string | null => {
    const name = line.name.toLowerCase();
    for (const b of sorted) {
      const bn = b.name.toLowerCase();
      if (name === bn || name.startsWith(bn + " ")) return b.name;
    }
    return null;
  };
}

export function Cart() {
  const { items, count, subtotalCents, isLoading, isError, removeItem, removingId, setQty, updatingId } =
    useCart();
  const { data: nav } = useNav();
  const resolveBrand = useMemo(() => makeBrandResolver(nav?.brands), [nav?.brands]);

  const crumbs = <Breadcrumb items={[{ label: "Home", to: "/" }, { label: "Cart" }]} />;

  if (isLoading) {
    return (
      <div className={styles.wrap}>
        {crumbs}
        <LoadingBlock label="Loading your cart…" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className={styles.wrap}>
        {crumbs}
        <ErrorBlock error={new Error("We couldn't load your cart. Please try again.")} />
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className={styles.wrap}>
        {crumbs}
        <div className={`state ${styles.empty}`}>
          <div className="state__title">Your cart is empty</div>
          <p className="state__body">Go find your ride — every build starts with one part.</p>
          <p className={styles.emptyCta}>
            <Link className="btn" to="/">
              Browse the shop
            </Link>
          </p>
        </div>
      </div>
    );
  }

  const freeShipping = subtotalCents >= FREE_SHIPPING_CENTS;
  const remaining = FREE_SHIPPING_CENTS - subtotalCents;

  return (
    <div className={styles.wrap}>
      {crumbs}

      <div className={styles.head}>
        <h1 className={styles.title}>Your cart</h1>
        <span className={styles.count}>
          {count} item{count === 1 ? "" : "s"}
        </span>
      </div>

      <div className={styles.layout}>
        <ul className={styles.lines} aria-label="Cart items">
          {items.map((line) => {
            const brand = resolveBrand(line);
            const updating = updatingId === line.productId;
            const busy = removingId === line.productId || updating;
            return (
              <li key={line.productId} className={styles.line} data-busy={busy}>
                <div className={styles.thumb}>
                  {line.slug ? (
                    <Link to={`/p/${line.slug}`} aria-label={line.name}>
                      <Thumb id={line.productId} kind={undefined} />
                    </Link>
                  ) : (
                    <Thumb id={line.productId} kind={undefined} />
                  )}
                </div>

                <div className={styles.lineInfo}>
                  {brand ? <div className={styles.brand}>{brand}</div> : null}
                  <div className={styles.name}>
                    {line.slug ? <Link to={`/p/${line.slug}`}>{line.name}</Link> : line.name}
                  </div>
                  <div className={styles.unit}>{formatCents(line.priceCents)} each</div>
                </div>

                <div className={styles.qty}>
                  <span className={styles.qtyLabel}>Qty</span>
                  <div className={styles.stepper}>
                    <button
                      type="button"
                      className={styles.step}
                      onClick={() => setQty(line.productId, line.qty - 1)}
                      disabled={busy}
                      aria-label={`Decrease quantity of ${line.name}`}
                    >
                      −
                    </button>
                    <span className={styles.qtyValue}>{line.qty}</span>
                    <button
                      type="button"
                      className={styles.step}
                      onClick={() => setQty(line.productId, line.qty + 1)}
                      disabled={busy}
                      aria-label={`Increase quantity of ${line.name}`}
                    >
                      +
                    </button>
                  </div>
                </div>

                <div className={styles.lineTotal}>{formatCents(line.priceCents * line.qty)}</div>

                <button
                  type="button"
                  className={styles.remove}
                  onClick={() => removeItem(line.productId)}
                  disabled={busy}
                  aria-label={`Remove ${line.name} from cart`}
                >
                  <CloseIcon />
                </button>
              </li>
            );
          })}
        </ul>

        <aside className={styles.summary} aria-label="Order summary">
          <h2 className={styles.summaryHead}>Order summary</h2>

          <dl className={styles.summaryRows}>
            <div className={styles.row}>
              <dt>Subtotal</dt>
              <dd className="mono">{formatCents(subtotalCents)}</dd>
            </div>
            <div className={styles.row}>
              <dt>Shipping</dt>
              <dd className="mono">{freeShipping ? "Free" : "Calculated at checkout"}</dd>
            </div>
          </dl>

          <p className={styles.shipNote}>
            {freeShipping
              ? "You've unlocked free shipping."
              : `Free shipping over $99 — you're ${formatCents(remaining)} away.`}
          </p>

          <div className={styles.totalRow}>
            <span>Estimated total</span>
            <span className={`mono ${styles.total}`}>{formatCents(subtotalCents)}</span>
          </div>

          <Link className={`btn btn--lg ${styles.checkout}`} to="/cart/checkout">
            Checkout
          </Link>
          <Link className={styles.keepShopping} to="/">
            Keep shopping
          </Link>
        </aside>
      </div>
    </div>
  );
}
