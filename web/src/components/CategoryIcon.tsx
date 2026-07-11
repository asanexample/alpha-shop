// Line-art icons, one per Kind — drawn as workshop blueprints (stroke only, no fill).
// Each icon is authored in a 200×120 viewBox; callers place it via x/y/width/height.
import type { SVGProps } from "react";
import type { Kind } from "../lib/types";

type IconProps = SVGProps<SVGSVGElement> & { kind: Kind };

// A few spokes for the wheel icons, generated so they're evenly spaced.
function spokes(cx: number, cy: number, r: number, n: number) {
  const lines = [];
  for (let i = 0; i < n; i++) {
    const a = (i / n) * Math.PI * 2;
    lines.push(
      <line key={i} x1={cx} y1={cy} x2={cx + Math.cos(a) * r} y2={cy + Math.sin(a) * r} />,
    );
  }
  return lines;
}

// Radial tread ticks around a tire.
function tread(cx: number, cy: number, rInner: number, rOuter: number, n: number) {
  const ticks = [];
  for (let i = 0; i < n; i++) {
    const a = (i / n) * Math.PI * 2;
    const c = Math.cos(a);
    const s = Math.sin(a);
    ticks.push(
      <line key={i} x1={cx + c * rInner} y1={cy + s * rInner} x2={cx + c * rOuter} y2={cy + s * rOuter} />,
    );
  }
  return ticks;
}

function Bike({ battery = false }: { battery?: boolean }) {
  return (
    <g>
      {/* wheels */}
      <circle cx={50} cy={84} r={26} />
      <circle cx={150} cy={84} r={26} />
      <circle cx={50} cy={84} r={2.5} className="dot" />
      <circle cx={150} cy={84} r={2.5} className="dot" />
      {/* diamond frame: BB(92,86) seat(82,46) head(126,46) */}
      <path d="M50 84 L92 86 L82 46 L50 84" />
      <path d="M92 86 L126 46 L82 46" />
      <path d="M92 86 L150 84" />
      {/* fork + head + bars */}
      <path d="M126 46 L150 84" />
      <path d="M126 46 L142 40 M142 40 L146 48" />
      {/* seat */}
      <path d="M74 45 L90 45" />
      {battery ? (
        <>
          <rect x={96} y={54} width={26} height={11} rx={2} className="accent-fill" />
          <path d="M133 40 l-8 12 h6 l-4 12" className="bolt" />
        </>
      ) : null}
    </g>
  );
}

export function CategoryIcon({ kind, ...svg }: IconProps) {
  return (
    <svg
      viewBox="0 0 200 120"
      role="img"
      aria-hidden="true"
      fill="none"
      stroke="currentColor"
      strokeWidth={3.4}
      strokeLinecap="round"
      strokeLinejoin="round"
      {...svg}
    >
      {kind === "bikes" && <Bike />}
      {kind === "ebikes" && <Bike battery />}

      {kind === "wheels" && (
        <g>
          <circle cx={100} cy={60} r={48} />
          <circle cx={100} cy={60} r={40} />
          <circle cx={100} cy={60} r={6} />
          {spokes(100, 60, 40, 10)}
        </g>
      )}

      {kind === "tires" && (
        <g>
          <circle cx={100} cy={60} r={50} />
          <circle cx={100} cy={60} r={34} />
          {tread(100, 60, 50, 58, 20)}
        </g>
      )}

      {kind === "components" && (
        <g>
          {/* chainring */}
          <circle cx={74} cy={62} r={34} />
          <circle cx={74} cy={62} r={12} />
          {tread(74, 62, 34, 40, 16)}
          {/* small cog */}
          <circle cx={150} cy={72} r={18} />
          <circle cx={150} cy={72} r={6} />
          {tread(150, 72, 18, 22, 12)}
          {/* chain line */}
          <path d="M74 96 Q112 108 150 90" strokeDasharray="1 7" />
        </g>
      )}

      {kind === "accessories" && (
        <g>
          {/* U-lock / padlock */}
          <rect x={64} y={58} width={72} height={50} rx={6} />
          <path d="M78 58 v-10 a22 22 0 0 1 44 0 v10" />
          <circle cx={100} cy={80} r={7} />
          <path d="M100 80 v14" />
        </g>
      )}

      {kind === "apparel" && (
        <g>
          {/* helmet dome */}
          <path d="M40 74 A62 62 0 0 1 164 74" />
          <path d="M40 74 q60 20 124 0" />
          {/* vents */}
          <path d="M74 54 l14 -12 M100 48 l14 -12 M126 52 l12 -10" />
          {/* strap */}
          <path d="M64 74 l10 20 M136 74 l-10 20" />
        </g>
      )}
    </svg>
  );
}
