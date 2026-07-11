import { Link } from "react-router-dom";
import { CategoryIcon } from "../components/CategoryIcon";
import styles from "./NotFound.module.css";

export function NotFound() {
  return (
    <div className={styles.wrap}>
      <div className={styles.mark} aria-hidden="true">
        4<span className={styles.zero}>0</span>4
      </div>
      <div>
        <p className={styles.code}>Dropped chain · Page not found</p>
        <h1 className={styles.title}>This route dead-ends</h1>
        <p className={styles.body}>
          The page you're after isn't here — maybe it moved, or the link was mistyped. Let's get you
          back on the road.
        </p>
        <div className={styles.actions}>
          <Link className="btn" to="/">
            Back to the shop
          </Link>
          <Link className="btn btn--ghost" to="/c/road">
            Shop bikes
          </Link>
        </div>
        <div style={{ marginTop: "2rem", width: 160, color: "var(--line-strong)" }}>
          <CategoryIcon kind="bikes" />
        </div>
      </div>
    </div>
  );
}
