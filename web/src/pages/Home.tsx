import { Link } from "react-router-dom";
import { CategoryIcon } from "../components/CategoryIcon";
import { ProductGrid } from "../components/ProductGrid";
import { ErrorBlock, LoadingBlock } from "../components/States";
import { useHome } from "../lib/hooks";
import { KIND_LABEL } from "../lib/types";
import styles from "./Home.module.css";

// Topographic / route-contour motif for the hero — layered wavy contour lines (a subtle map feel).
function TopoLines() {
  const rows = 9;
  const paths: string[] = [];
  for (let r = 0; r < rows; r++) {
    const y = 40 + r * 46;
    const amp = 14 + (r % 3) * 8;
    const seg = 160;
    let d = `M-20 ${y}`;
    for (let x = 0; x <= 1280; x += seg) {
      const dir = (Math.floor(x / seg) + r) % 2 === 0 ? -1 : 1;
      d += ` q ${seg / 2} ${dir * amp} ${seg} 0`;
    }
    paths.push(d);
  }
  return (
    <svg className={styles.topo} viewBox="0 0 1240 460" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
      {paths.map((d, i) => (
        <path key={i} d={d} fill="none" stroke="currentColor" strokeWidth={1.25} />
      ))}
    </svg>
  );
}

const SERVICES = [
  { name: "Standard tune-up", note: "Brakes, shifting, safety check — next-day turnaround.", price: "$85" },
  { name: "Custom wheel build", note: "Hand-laced, tensioned, and trued to your ride.", price: "from $180" },
  { name: "Full frame-up build", note: "Bring a frame or a dream; we'll spec and assemble it.", price: "quote" },
  { name: "Bike fit session", note: "60 minutes on the fit jig with a certified fitter.", price: "$120" },
];

const COMMUNITY = [
  {
    tag: "Every Saturday",
    title: "Shop Ride",
    body: "A no-drop 25-mile spin from the front door at 8am. Coffee after, always.",
  },
  {
    tag: "First Tuesday",
    title: "Wrench Night",
    body: "Free, hands-on repair classes. Bring your bike and learn to fix it yourself.",
  },
  {
    tag: "Since 2009",
    title: "Trail Advocacy",
    body: "We put a share of every sale toward local trail building and safe-streets work.",
  },
];

export function Home() {
  const { data, isLoading, isError, error, refetch } = useHome();

  return (
    <div>
      {/* ---- HERO ---- */}
      <section className={styles.hero} aria-labelledby="hero-title">
        <TopoLines />
        <div className={styles.heroInner}>
          <p className={styles.heroEyebrow}>Portland · Independent · Since 2009</p>
          <h1 id="hero-title" className={styles.heroTitle}>
            Built to ride.
            <br />
            <span className={styles.em}>Tuned by hand.</span>
          </h1>
          <p className={styles.heroSub}>
            Road, gravel, mountain, and electric — every bike checked over, dialed in, and ready to
            roll out the door. Real mechanics, honest advice, no showroom nonsense.
          </p>
          <div className={styles.heroActions}>
            <Link className="btn btn--lg" to="/c/road">
              Shop bikes
            </Link>
            <Link className={`btn btn--lg ${styles.ghostOnInk}`} to="/service">
              Book service
            </Link>
          </div>
          <dl className={styles.heroStats}>
            <div>
              <b>15 yrs</b> in the neighborhood
            </div>
            <div>
              <b>7 brands</b> we ride ourselves
            </div>
            <div>
              <b>Next-day</b> tune-ups
            </div>
          </dl>
        </div>
      </section>

      {/* ---- SHOP BY CATEGORY ---- */}
      <section className={styles.section} aria-labelledby="cat-title">
        <div className="section-head">
          <div>
            <p className="eyebrow">Find your ride</p>
            <h2 id="cat-title" className="section-title">
              Shop by category
            </h2>
          </div>
        </div>

        {isLoading ? (
          <LoadingBlock label="Loading the shop…" />
        ) : isError ? (
          <ErrorBlock error={error} onRetry={() => refetch()} />
        ) : (
          <div className={styles.tiles}>
            {data?.categories.map((c) => (
              <Link key={c.slug} to={`/c/${c.slug}`} className={styles.tile}>
                <span className={styles.tileArrow} aria-hidden="true">
                  →
                </span>
                <CategoryIcon kind={c.kind} className={styles.tileIcon} />
                <span className={styles.tileKind}>{KIND_LABEL[c.kind]}</span>
                <span className={styles.tileName}>{c.name}</span>
                {c.blurb ? <span className={styles.tileBlurb}>{c.blurb}</span> : null}
              </Link>
            ))}
          </div>
        )}
      </section>

      {/* ---- FEATURED ---- */}
      <section className={styles.section} aria-labelledby="feat-title">
        <div className="section-head">
          <div>
            <p className="eyebrow">Picked by the mechanics</p>
            <h2 id="feat-title" className="section-title">
              Featured this week
            </h2>
          </div>
          <Link className="btn btn--ghost" to="/search?q=">
            Browse everything
          </Link>
        </div>

        {isLoading ? (
          <LoadingBlock />
        ) : isError ? (
          <ErrorBlock error={error} onRetry={() => refetch()} />
        ) : data && data.featured.length > 0 ? (
          <ProductGrid products={data.featured} />
        ) : (
          <p className="mono">Nothing featured right now — check back soon.</p>
        )}
      </section>

      {/* ---- SERVICE & CUSTOM BUILDS ---- */}
      <section className={styles.service} aria-labelledby="svc-title">
        <div className={styles.serviceInner}>
          <div>
            <p className="eyebrow">The workshop</p>
            <h2 id="svc-title" className={styles.serviceTitle}>
              Service &amp; custom builds
            </h2>
            <p className={styles.serviceBody}>
              The shop started as a repair bench, and it still is one. Whatever you ride — bought
              here or not — our mechanics will get it running right. Book online or just roll in.
            </p>
            <p style={{ marginTop: "1.5rem" }}>
              <Link className="btn" to="/service">
                Book a service
              </Link>
            </p>
          </div>
          <div className={styles.serviceList}>
            {SERVICES.map((s) => (
              <div key={s.name} className={styles.serviceItem}>
                <div>
                  <h3>{s.name}</h3>
                  <p>{s.note}</p>
                </div>
                <span className={styles.servicePrice}>{s.price}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ---- COMMUNITY / HERITAGE ---- */}
      <section className={styles.community} aria-labelledby="comm-title">
        <div className="section-head">
          <div>
            <p className="eyebrow">More than a store</p>
            <h2 id="comm-title" className="section-title">
              Rooted in the ride
            </h2>
          </div>
        </div>
        <div className={styles.communityGrid}>
          {COMMUNITY.map((c) => (
            <article key={c.title} className={styles.commCard}>
              <span className={styles.commNum}>{c.tag}</span>
              <h3>{c.title}</h3>
              <p>{c.body}</p>
            </article>
          ))}
        </div>
      </section>
    </div>
  );
}
