import { useQuery } from "@tanstack/react-query";
import { Link, Navigate } from "react-router-dom";
import { Breadcrumb } from "../components/Breadcrumb";
import { useAuth } from "../context/AuthContext";
import { ErrorBlock, LoadingBlock } from "../components/States";
import { api } from "../lib/api";
import { formatCents, formatDate } from "../lib/format";
import styles from "./Orders.module.css";

export function Orders() {
  const { isAuthenticated, isLoading: authLoading } = useAuth();

  const ordersQuery = useQuery({
    queryKey: ["orders"],
    queryFn: ({ signal }) => api.orders(signal),
    enabled: isAuthenticated,
  });

  const crumbs = <Breadcrumb items={[{ label: "Home", to: "/" }, { label: "Your orders" }]} />;

  if (authLoading) {
    return (
      <div className={styles.wrap}>
        {crumbs}
        <LoadingBlock label="Loading your account…" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login?next=/account/orders" replace />;
  }

  const { data, isLoading, isError, error, refetch } = ordersQuery;
  const orders = data?.orders ?? [];

  return (
    <div className={styles.wrap}>
      {crumbs}

      <h1 className={styles.title}>Your orders</h1>

      {isLoading ? (
        <LoadingBlock label="Loading your orders…" />
      ) : isError ? (
        <ErrorBlock error={error} onRetry={() => refetch()} />
      ) : orders.length === 0 ? (
        <div className={`state ${styles.empty}`}>
          <div className="state__title">No orders yet</div>
          <p className="state__body">Once you check out, your orders will show up here.</p>
          <p className={styles.emptyCta}>
            <Link className="btn" to="/">
              Browse the shop
            </Link>
          </p>
        </div>
      ) : (
        <ul className={styles.list} aria-label="Order history">
          {orders.map((o) => (
            <li key={o.id}>
              <Link to={`/order/${o.id}`} className={styles.row} data-declined={o.status === "declined"}>
                <span className={`mono ${styles.id}`}>{o.id}</span>
                <span className={styles.date}>{formatDate(o.createdAt)}</span>
                <span className={styles.status}>{o.status}</span>
                <span className={`mono ${styles.total}`}>{formatCents(o.totalCents)}</span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
