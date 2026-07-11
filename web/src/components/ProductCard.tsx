import { Link } from "react-router-dom";
import { formatCents } from "../lib/format";
import { useNavLookups } from "../lib/hooks";
import { isOnSale, type Product } from "../lib/types";
import { Thumb } from "./Thumb";
import styles from "./ProductCard.module.css";

export function ProductCard({ product }: { product: Product }) {
  const { brandName, categoryKind } = useNavLookups();
  const sale = isOnSale(product);

  return (
    <article className={styles.card}>
      <Link className={styles.link} to={`/p/${product.slug}`}>
        <div className={styles.media}>
          <div className={styles.badges}>
            {sale ? <span className="tag tag--sale">Sale</span> : null}
            {!product.inStock ? <span className="tag tag--out">Sold out</span> : null}
          </div>
          <Thumb id={product.id} kind={categoryKind(product.category)} />
        </div>
        <div className={styles.body}>
          <div className={styles.brand}>{brandName(product.brand)}</div>
          <h3 className={styles.name}>{product.name}</h3>
          <p className={styles.summary}>{product.summary}</p>
          <div className={styles.priceRow}>
            {sale ? (
              <>
                <span className={`${styles.price} ${styles["price--sale"]}`}>
                  {formatCents(product.salePriceCents as number)}
                </span>
                <span className={styles.list}>{formatCents(product.priceCents)}</span>
              </>
            ) : (
              <span className={styles.price}>{formatCents(product.priceCents)}</span>
            )}
            {!product.inStock ? <span className={styles.outNote}>Out of stock</span> : null}
          </div>
        </div>
      </Link>
    </article>
  );
}
