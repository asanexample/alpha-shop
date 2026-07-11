import { Fragment } from "react";
import { Link } from "react-router-dom";
import styles from "./Breadcrumb.module.css";

export interface Crumb {
  label: string;
  to?: string;
}

export function Breadcrumb({ items }: { items: Crumb[] }) {
  return (
    <nav className={styles.crumb} aria-label="Breadcrumb">
      {items.map((item, i) => {
        const last = i === items.length - 1;
        return (
          <Fragment key={i}>
            {item.to && !last ? (
              <Link to={item.to}>{item.label}</Link>
            ) : (
              <span className={styles.current} aria-current={last ? "page" : undefined}>
                {item.label}
              </span>
            )}
            {!last ? <span className={styles.sep} aria-hidden="true">/</span> : null}
          </Fragment>
        );
      })}
    </nav>
  );
}
