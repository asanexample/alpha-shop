import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Breadcrumb } from "../components/Breadcrumb";
import { ProductGrid } from "../components/ProductGrid";
import { ErrorBlock, LoadingBlock } from "../components/States";
import { Thumb } from "../components/Thumb";
import { useCart } from "../context/CartContext";
import { ApiError } from "../lib/api";
import { formatCents, savingsLabel } from "../lib/format";
import { useNavLookups, useProduct } from "../lib/hooks";
import { isOnSale, KIND_LABEL } from "../lib/types";
import styles from "./Product.module.css";

export function ProductDetail() {
  const { slug } = useParams<{ slug: string }>();
  const { data, isLoading, isError, error, refetch } = useProduct(slug);
  const { brandName, categoryName, categoryKind, category } = useNavLookups();
  const { add } = useCart();

  const [size, setSize] = useState<string | null>(null);
  // Reset the selected size when the product changes.
  useEffect(() => setSize(null), [slug]);

  if (isLoading) {
    return (
      <div className={styles.wrap}>
        <LoadingBlock label="Loading product…" />
      </div>
    );
  }

  if (isError) {
    const notFound = error instanceof ApiError && error.status === 404;
    return (
      <div className={styles.wrap}>
        {notFound ? (
          <div className="state">
            <div className="state__title">We can't find that bike</div>
            <p className="state__body">
              This product may have sold out or moved. Browse the shop to find your ride.
            </p>
            <p style={{ marginTop: "1.25rem" }}>
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

  if (!data) return null;
  const { product, related } = data;
  const sale = isOnSale(product);
  const cat = category(product.category);
  const sizes = product.sizes ?? [];

  function handleAdd() {
    const label = size ? `${product.name} · ${size}` : product.name;
    add(label);
  }

  return (
    <div className={styles.wrap}>
      <Breadcrumb
        items={[
          { label: "Home", to: "/" },
          { label: cat ? KIND_LABEL[cat.kind] : "Shop" },
          { label: categoryName(product.category), to: `/c/${product.category}` },
          { label: product.name },
        ]}
      />

      <div className={styles.layout}>
        <div className={styles.media}>
          <Thumb id={product.id} kind={categoryKind(product.category)} sku={product.id} />
        </div>

        <div className={styles.info}>
          <p className={styles.brand}>{brandName(product.brand)}</p>
          <h1 className={styles.name}>{product.name}</h1>
          <p className={styles.summary}>{product.summary}</p>

          <div className={styles.priceRow}>
            {sale ? (
              <>
                <span className={styles.price}>{formatCents(product.salePriceCents as number)}</span>
                <span className={styles.list}>{formatCents(product.priceCents)}</span>
                <span className={styles.saveTag}>
                  Save {savingsLabel(product.priceCents, product.salePriceCents as number).replace("-", "")}
                </span>
              </>
            ) : (
              <span className={styles.price}>{formatCents(product.priceCents)}</span>
            )}
          </div>

          <div className={`${styles.stock} ${product.inStock ? styles.inStock : styles.outStock}`}>
            <span className={styles.dot} aria-hidden="true" />
            {product.inStock ? "In stock — ships in 1–2 days" : "Out of stock — join the waitlist"}
          </div>

          {sizes.length > 0 ? (
            <div className={styles.sizes}>
              <div className={styles.sizesLabel}>Size</div>
              <div className={styles.chips} role="group" aria-label="Select a size">
                {sizes.map((s) => (
                  <button
                    key={s}
                    type="button"
                    className={styles.chip}
                    data-selected={size === s}
                    aria-pressed={size === s}
                    onClick={() => setSize((cur) => (cur === s ? null : s))}
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>
          ) : null}

          <div className={styles.buyRow}>
            <button
              type="button"
              className={`btn btn--lg ${styles.add}`}
              disabled={!product.inStock}
              onClick={handleAdd}
            >
              {product.inStock ? "Add to cart" : "Sold out"}
            </button>
          </div>

          {product.specs && Object.keys(product.specs).length > 0 ? (
            <div className={styles.specs}>
              <div className={styles.specsHead}>Specifications</div>
              <table className={styles.specsTable}>
                <tbody>
                  {Object.entries(product.specs).map(([k, v]) => (
                    <tr key={k}>
                      <th scope="row">{k}</th>
                      <td>{v}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : null}

          <h2 className={styles.descHead}>The details</h2>
          <p className={styles.desc}>{product.description}</p>
        </div>
      </div>

      {related.length > 0 ? (
        <section className={styles.related} aria-labelledby="related-title">
          <div className="section-head">
            <div>
              <p className="eyebrow">In the same category</p>
              <h2 id="related-title" className="section-title">
                You might also like
              </h2>
            </div>
          </div>
          <ProductGrid products={related} />
        </section>
      ) : null}
    </div>
  );
}
