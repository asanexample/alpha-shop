import { useCart } from "../context/CartContext";
import styles from "./Toaster.module.css";

export function Toaster() {
  const { toasts, dismiss } = useCart();
  if (toasts.length === 0) return null;
  return (
    <div className={styles.wrap} aria-live="polite" aria-atomic="false">
      {toasts.map((t) => (
        <div key={t.id} className={styles.toast} role="status">
          <svg className={styles.icon} viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <path d="M4 12.5 9 17.5 20 6.5" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
          <span className={styles.msg}>{t.message}</span>
          <button type="button" className={styles.close} onClick={() => dismiss(t.id)} aria-label="Dismiss">
            ×
          </button>
        </div>
      ))}
    </div>
  );
}
