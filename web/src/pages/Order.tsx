import { useQuery } from "@tanstack/react-query";
import { Link, useParams } from "react-router-dom";
import { ErrorBlock, LoadingBlock } from "../components/States";
import { ApiError, api } from "../lib/api";
import { formatCents } from "../lib/format";
import styles from "./Order.module.css";

export function Order() {
  const { id } = useParams<{ id: string }>();
  const { data: order, isLoading, isError, error, refetch } = useQuery({
    queryKey: ["order", id],
    queryFn: ({ signal }) => api.order(id as string, signal),
    enabled: !!id,
    retry: (count, err) => ((err as { status?: number })?.status === 404 ? false : count < 2),
  });

  if (isLoading) {
    return (
      <div className={styles.wrap}>
        <LoadingBlock label="Loading your order…" />
      </div>
    );
  }

  if (isError) {
    const notFound = error instanceof ApiError && error.status === 404;
    return (
      <div className={styles.wrap}>
        {notFound ? (
          <div className="state">
            <div className="state__title">We can't find that order</div>
            <p className="state__body">
              This order number doesn't match anything in the workshop. Check the link and try again.
            </p>
            <p className={styles.cta}>
              <Link className="btn" to="/">
                Back to the shop
              </Link>
            </p>
          </div>
        ) : (
          <ErrorBlock error={error} onRetry={() => refetch()} />
        )}
      </div>
    );
  }

  if (!order) return null;
  const declined = order.status === "declined";

  return (
    <div className={styles.wrap}>
      <header className={styles.hero} data-declined={declined}>
        <span className={styles.mark} aria-hidden="true">
          {declined ? (
            <svg viewBox="0 0 48 48" className={styles.markSvg}>
              <line x1="15" y1="15" x2="33" y2="33" />
              <line x1="33" y1="15" x2="15" y2="33" />
            </svg>
          ) : (
            <svg viewBox="0 0 48 48" className={styles.markSvg}>
              <path d="M13 25 21 33 36 16" />
            </svg>
          )}
        </span>

        <p className={styles.eyebrow}>{declined ? "Payment declined" : "Thank you"}</p>
        <h1 className={styles.title}>{declined ? "Order not placed" : "Order placed"}</h1>
        <p className={styles.lede}>
          {declined
            ? order.reason || "Your card was declined and you haven't been charged."
            : "Your order is confirmed. We're prepping it in the workshop — you'll get a shipping note soon."}
        </p>
      </header>

      <div className={styles.meta}>
        <div className={styles.metaItem}>
          <span className={styles.metaLabel}>Order</span>
          <span className={`mono ${styles.metaValue}`}>{order.id}</span>
        </div>
        <div className={styles.metaItem}>
          <span className={styles.metaLabel}>Payment</span>
          <span className={`mono ${styles.metaValue}`}>{order.paymentId || "—"}</span>
        </div>
        <div className={styles.metaItem}>
          <span className={styles.metaLabel}>Status</span>
          <span className={`mono ${styles.metaValue}`}>{order.status}</span>
        </div>
      </div>

      <section className={styles.card} aria-label="Order items">
        <h2 className={styles.cardHead}>
          {declined ? "Attempted items" : "What's on the way"}
        </h2>
        <ul className={styles.lines}>
          {order.lines.map((line) => (
            <li key={line.productId} className={styles.line}>
              <span className={styles.lineQty}>{line.qty}×</span>
              <span className={styles.lineName}>
                {line.slug ? <Link to={`/p/${line.slug}`}>{line.name}</Link> : line.name}
              </span>
              <span className={styles.lineUnit}>{formatCents(line.priceCents)} ea</span>
              <span className={`mono ${styles.linePrice}`}>
                {formatCents(line.priceCents * line.qty)}
              </span>
            </li>
          ))}
        </ul>
        {!declined && order.shipping ? (
          <div className={styles.shipRow}>
            <span>
              Shipping
              {order.experience === "express" ? (
                <span className={styles.expressBadge}>EXPRESS</span>
              ) : null}
            </span>
            <span className={styles.shipValue}>{order.shipping}</span>
          </div>
        ) : null}
        <div className={styles.totalRow}>
          <span>{declined ? "Amount" : "Total"}</span>
          <span className={`mono ${styles.total}`}>{formatCents(order.totalCents)}</span>
        </div>
      </section>

      <div className={styles.actions}>
        {declined ? (
          <Link className="btn btn--lg" to="/cart/checkout">
            Try another card
          </Link>
        ) : null}
        <Link className={declined ? "btn btn--lg btn--ghost" : "btn btn--lg"} to="/">
          Continue shopping
        </Link>
      </div>
    </div>
  );
}
