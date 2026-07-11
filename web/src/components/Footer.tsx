import { Link } from "react-router-dom";
import { Chainring } from "./Glyphs";
import styles from "./Footer.module.css";

const COLUMNS: { title: string; links: { label: string; to: string }[] }[] = [
  {
    title: "Shop",
    links: [
      { label: "Road Bikes", to: "/c/road" },
      { label: "Gravel Bikes", to: "/c/gravel" },
      { label: "Mountain Bikes", to: "/c/mountain" },
      { label: "E-Bikes", to: "/c/ebike-city" },
      { label: "Parts & Tires", to: "/c/drivetrain" },
    ],
  },
  {
    title: "Service",
    links: [
      { label: "Book a tune-up", to: "/service" },
      { label: "Custom builds", to: "/service" },
      { label: "Wheel building", to: "/service" },
      { label: "Bike fitting", to: "/service" },
    ],
  },
  {
    title: "About",
    links: [
      { label: "Our story", to: "/about" },
      { label: "Visit the shop", to: "/about" },
      { label: "Careers", to: "/about" },
      { label: "Returns & warranty", to: "/about" },
    ],
  },
  {
    title: "Community",
    links: [
      { label: "Group rides", to: "/community" },
      { label: "Repair classes", to: "/community" },
      { label: "Trail advocacy", to: "/community" },
      { label: "Ride journal", to: "/community" },
    ],
  },
];

export function Footer() {
  return (
    <footer className={styles.footer}>
      <div className={styles.inner}>
        <div className={styles.top}>
          <div className={styles.brandCol}>
            <Link to="/" className={styles.wordmark} aria-label="Alpha Bikes — home">
              <Chainring className={styles.ring} /> Alpha Bikes
            </Link>
            <p className={styles.tagline}>
              A full-service bike shop on the east side of Portland. We sell bikes we ride and fix
              bikes we sell — no gatekeeping, no upsell.
            </p>
            <p className={styles.heritage}>Independent &amp; rider-owned since 2009</p>
          </div>

          {COLUMNS.map((col) => (
            <nav key={col.title} className={styles.col} aria-label={col.title}>
              <div className={styles.colTitle}>{col.title}</div>
              {col.links.map((l) => (
                <Link key={l.label} to={l.to}>
                  {l.label}
                </Link>
              ))}
            </nav>
          ))}
        </div>

        <div className={styles.bottom}>
          <span>© {new Date().getFullYear()} Alpha Bikes Co-op · 2340 SE Belmont St, Portland OR</span>
          <span>
            <Link to="/about">Privacy</Link> · <Link to="/about">Terms</Link> · Mon–Sat 10–7
          </span>
        </div>
      </div>
    </footer>
  );
}
