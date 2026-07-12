import { useState, type FormEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate } from "react-router-dom";
import { Breadcrumb } from "../components/Breadcrumb";
import { LoadingBlock } from "../components/States";
import { CART_QUERY_KEY, useCart } from "../context/CartContext";
import { api } from "../lib/api";
import { formatCents } from "../lib/format";
import type { Order } from "../lib/types";
import styles from "./Checkout.module.css";

export function Checkout() {
  const { items, subtotalCents, count, isLoading } = useCart();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [card, setCard] = useState("");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [declined, setDeclined] = useState<string | null>(null);

  const placeOrder = useMutation({
    mutationFn: () => api.checkout(card.trim()),
    onSuccess: (order: Order) => {
      if (order.status === "placed") {
        // The BFF clears the cart on a placed order — reflect that locally.
        void queryClient.invalidateQueries({ queryKey: CART_QUERY_KEY });
        navigate(`/order/${order.id}`);
        return;
      }
      // Declined: stay here so they can retry with another card.
      setDeclined(order.reason || "Your card was declined. Please try another card.");
    },
    onError: (err) => {
      setDeclined(err instanceof Error ? err.message : "Something went wrong placing your order.");
    },
  });

  const crumbs = (
    <Breadcrumb
      items={[{ label: "Home", to: "/" }, { label: "Cart", to: "/cart" }, { label: "Checkout" }]}
    />
  );

  if (isLoading) {
    return (
      <div className={styles.wrap}>
        {crumbs}
        <LoadingBlock label="Loading checkout…" />
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div className={styles.wrap}>
        {crumbs}
        <div className={`state ${styles.empty}`}>
          <div className="state__title">Nothing to check out</div>
          <p className="state__body">Your cart is empty. Add a part or a bike to get rolling.</p>
          <p className={styles.emptyCta}>
            <Link className="btn" to="/">
              Browse the shop
            </Link>
          </p>
        </div>
      </div>
    );
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    setDeclined(null);
    placeOrder.mutate();
  }

  const placing = placeOrder.isPending;

  return (
    <div className={styles.wrap}>
      {crumbs}

      <h1 className={styles.title}>Checkout</h1>

      <div className={styles.layout}>
        <form className={styles.form} onSubmit={onSubmit} noValidate>
          <fieldset className={styles.fieldset} disabled={placing}>
            <legend className={styles.legend}>Contact</legend>
            <label className={styles.field}>
              <span className={styles.label}>Name</span>
              <input
                className={styles.input}
                type="text"
                autoComplete="name"
                placeholder="Alex Rider"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </label>
            <label className={styles.field}>
              <span className={styles.label}>Email</span>
              <input
                className={styles.input}
                type="email"
                autoComplete="email"
                placeholder="you@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </label>
          </fieldset>

          <fieldset className={styles.fieldset} disabled={placing}>
            <legend className={styles.legend}>Payment</legend>
            <label className={styles.field}>
              <span className={styles.label}>Card</span>
              <input
                className={`${styles.input} mono`}
                type="text"
                inputMode="numeric"
                autoComplete="cc-number"
                placeholder="4242 4242 4242 4242"
                value={card}
                onChange={(e) => setCard(e.target.value)}
                aria-describedby="card-help"
              />
            </label>
            <p id="card-help" className={styles.help}>
              Demo only — no real charge. Tip: a card ending in 0000 simulates a decline.
            </p>
          </fieldset>

          {declined ? (
            <div className={styles.declined} role="alert">
              <strong>Payment declined.</strong> {declined}
            </div>
          ) : null}

          <button type="submit" className={`btn btn--lg ${styles.place}`} disabled={placing}>
            {placing ? "Placing order…" : `Place order · ${formatCents(subtotalCents)}`}
          </button>
        </form>

        <aside className={styles.summary} aria-label="Order summary">
          <h2 className={styles.summaryHead}>
            Order summary
            <span className={styles.summaryCount}>
              {count} item{count === 1 ? "" : "s"}
            </span>
          </h2>

          <ul className={styles.items}>
            {items.map((line) => (
              <li key={line.productId} className={styles.item}>
                <span className={styles.itemQty}>{line.qty}×</span>
                <span className={styles.itemName}>{line.name}</span>
                <span className={`mono ${styles.itemPrice}`}>
                  {formatCents(line.priceCents * line.qty)}
                </span>
              </li>
            ))}
          </ul>

          <div className={styles.totalRow}>
            <span>Subtotal</span>
            <span className={`mono ${styles.total}`}>{formatCents(subtotalCents)}</span>
          </div>
          <p className={styles.shipNote}>Shipping calculated after payment · Free over $99.</p>
        </aside>
      </div>
    </div>
  );
}
