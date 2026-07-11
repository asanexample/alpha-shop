// The visual identity: a deterministic "workshop blueprint" swatch standing in for a product photo.
// Graph-paper grid + a Kind line-icon + a hue derived from the product id + a hint of the yellow accent.
import { useId } from "react";
import { blueprintFor } from "../lib/hue";
import type { Kind } from "../lib/types";
import { CategoryIcon } from "./CategoryIcon";
import styles from "./Thumb.module.css";

interface ThumbProps {
  id: string;
  kind: Kind | undefined;
  sku?: string; // small mono label (defaults to the product id)
  className?: string;
}

const W = 400;
const H = 300;

export function Thumb({ id, kind, sku, className }: ThumbProps) {
  const bp = blueprintFor(id);
  const uid = useId().replace(/:/g, "");
  const fineId = `grid-${uid}`;
  const boldId = `gridbold-${uid}`;
  const label = (sku ?? id).toUpperCase();

  return (
    <div
      className={`${styles.thumb} ${className ?? ""}`}
      style={{ ["--icon-ink" as string]: bp.ink, background: bp.paper }}
    >
      <svg
        className={styles.svg}
        viewBox={`0 0 ${W} ${H}`}
        preserveAspectRatio="xMidYMid slice"
        aria-hidden="true"
      >
        <defs>
          <pattern id={fineId} width="20" height="20" patternUnits="userSpaceOnUse">
            <path d="M20 0 H0 V20" fill="none" stroke={bp.grid} strokeWidth="1" />
          </pattern>
          <pattern id={boldId} width="100" height="100" patternUnits="userSpaceOnUse">
            <path d="M100 0 H0 V100" fill="none" stroke={bp.gridBold} strokeWidth="1.25" />
          </pattern>
        </defs>

        {/* graph paper */}
        <rect x="0" y="0" width={W} height={H} fill={bp.paper} />
        <rect x="0" y="0" width={W} height={H} fill={`url(#${fineId})`} />
        <rect x="0" y="0" width={W} height={H} fill={`url(#${boldId})`} />

        {/* drafting frame */}
        <rect x="10" y="10" width={W - 20} height={H - 20} fill="none" stroke={bp.ink} strokeOpacity="0.32" strokeWidth="1.25" />

        {/* the line-icon, centred */}
        <CategoryIcon
          kind={kind ?? "bikes"}
          className={styles.icon}
          x={90}
          y={78}
          width={220}
          height={132}
        />

        {/* yellow registration cross — the single accent hint */}
        <g stroke="var(--accent)" strokeWidth="2.5">
          <line x1="356" y1="34" x2="356" y2="54" />
          <line x1="346" y1="44" x2="366" y2="44" />
        </g>

        {/* mono SKU label, blueprint-style */}
        <text className={styles.label} x="22" y="40">
          {label}
        </text>

        {/* little drafting scale bar */}
        <g>
          <line x1="22" y1="272" x2="82" y2="272" stroke={bp.ink} strokeOpacity="0.4" strokeWidth="1.5" />
          <line x1="22" y1="268" x2="22" y2="276" stroke={bp.ink} strokeOpacity="0.4" strokeWidth="1.5" />
          <line x1="82" y1="268" x2="82" y2="276" stroke={bp.ink} strokeOpacity="0.4" strokeWidth="1.5" />
          <text className={styles.scaleText} x="90" y="276">
            SCALE 1:1
          </text>
        </g>
      </svg>
    </div>
  );
}
