import { useState, type FormEvent } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { Breadcrumb } from "../components/Breadcrumb";
import { LoadingBlock } from "../components/States";
import { useAuth } from "../context/AuthContext";
import { CART_QUERY_KEY, useCart } from "../context/CartContext";
import { api } from "../lib/api";
import { formatCents } from "../lib/format";
import { PAYMENT_METHOD_LABEL, type Address, type Order, type PaymentMethod } from "../lib/types";
import styles from "./Checkout.module.css";

const PAYMENT_METHODS: PaymentMethod[] = ["card", "paypal", "apple_pay"];

export function Checkout() {
  const { isAuthenticated, isLoading: authLoading, user } = useAuth();
  const { items, subtotalCents, count, isLoading: cartLoading } = useCart();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [card, setCard] = useState("");
  const [method, setMethod] = useState<PaymentMethod>("card");
  const [name, setName] = useState("");
  const [line1, setLine1] = useState("");
  const [line2, setLine2] = useState("");
  const [city, setCity] = useState("");
  const [state, setState] = useState("");
  const [zip, setZip] = useState("");
  const [country, setCountry] = useState("United States");
  const [declined, setDeclined] = useState<string | null>(null);

  const placeOrder = useMutation({
    mutationFn: () => {
      const address: Address = {
        name: name.trim(),
        line1: line1.trim(),
        line2: line2.trim() || undefined,
        city: city.trim(),
        state: state.trim(),
        zip: zip.trim(),
        country: country.trim(),
      };
      return api.checkout({ card: card.trim(), address, paymentMethod: method });
    },
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

  if (authLoading || cartLoading) {
    return (
      <div className={styles.wrap}>
        {crumbs}
        <LoadingBlock label="Loading checkout…" />
      </div>
    );
  }

  // Order history has no stable identity to key on for a guest, so checkout is the point signup/login
  // becomes mandatory — send them to sign in first, then straight back here.
  if (!isAuthenticated) {
    return <Navigate to="/login?next=/cart/checkout" replace />;
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
            <legend className={styles.legend}>Signed in</legend>
            <p className={styles.help}>
              {user?.name || user?.email} · <Link to="/account/orders">View your orders</Link>
            </p>
          </fieldset>

          <fieldset className={styles.fieldset} disabled={placing}>
            <legend className={styles.legend}>Shipping address</legend>
            <label className={styles.field}>
              <span className={styles.label}>Full name</span>
              <input
                className={styles.input}
                type="text"
                autoComplete="name"
                required
                placeholder="Alex Rider"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </label>
            <label className={styles.field}>
              <span className={styles.label}>Address</span>
              <input
                className={styles.input}
                type="text"
                autoComplete="address-line1"
                required
                placeholder="123 Workshop Way"
                value={line1}
                onChange={(e) => setLine1(e.target.value)}
              />
            </label>
            <label className={styles.field}>
              <span className={styles.label}>Apartment, suite, etc. (optional)</span>
              <input
                className={styles.input}
                type="text"
                autoComplete="address-line2"
                value={line2}
                onChange={(e) => setLine2(e.target.value)}
              />
            </label>
            <div className={styles.addrGrid}>
              <label className={styles.field}>
                <span className={styles.label}>City</span>
                <input
                  className={styles.input}
                  type="text"
                  autoComplete="address-level2"
                  required
                  value={city}
                  onChange={(e) => setCity(e.target.value)}
                />
              </label>
              <label className={styles.field}>
                <span className={styles.label}>State</span>
                <input
                  className={styles.input}
                  type="text"
                  autoComplete="address-level1"
                  required
                  placeholder="OR"
                  value={state}
                  onChange={(e) => setState(e.target.value)}
                />
              </label>
              <label className={styles.field}>
                <span className={styles.label}>ZIP</span>
                <input
                  className={styles.input}
                  type="text"
                  inputMode="numeric"
                  autoComplete="postal-code"
                  required
                  value={zip}
                  onChange={(e) => setZip(e.target.value)}
                />
              </label>
            </div>
            <label className={styles.field}>
              <span className={styles.label}>Country</span>
              <input
                className={styles.input}
                type="text"
                autoComplete="country-name"
                required
                value={country}
                onChange={(e) => setCountry(e.target.value)}
              />
            </label>
          </fieldset>

          <fieldset className={styles.fieldset} disabled={placing}>
            <legend className={styles.legend}>Payment</legend>
            <div className={styles.methodGroup} role="radiogroup" aria-label="Payment method">
              {PAYMENT_METHODS.map((m) => (
                <label key={m} className={styles.methodOption} data-active={method === m}>
                  <input
                    type="radio"
                    name="paymentMethod"
                    value={m}
                    checked={method === m}
                    onChange={() => setMethod(m)}
                  />
                  {PAYMENT_METHOD_LABEL[m]}
                </label>
              ))}
            </div>
            {method === "card" ? (
              <>
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
              </>
            ) : (
              <p className={styles.help}>
                Demo only — no real charge. You'll be "redirected" nowhere; placing the order simulates
                the {PAYMENT_METHOD_LABEL[method].toLowerCase()} flow.
              </p>
            )}
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
