// Small inline SVG glyphs used in the chrome (wordmark chainring, search, cart, menu).
import type { SVGProps } from "react";

export function Chainring(props: SVGProps<SVGSVGElement>) {
  // A stylised chainring — the wordmark mark.
  const teeth = [];
  for (let i = 0; i < 12; i++) {
    const a = (i / 12) * Math.PI * 2;
    teeth.push(
      <rect
        key={i}
        x={11}
        y={0.5}
        width={2}
        height={3}
        rx={0.5}
        transform={`rotate(${(a * 180) / Math.PI} 12 12)`}
      />,
    );
  }
  return (
    <svg viewBox="0 0 24 24" width="1em" height="1em" aria-hidden="true" {...props}>
      <g fill="currentColor">{teeth}</g>
      <circle cx={12} cy={12} r={8.5} fill="none" stroke="currentColor" strokeWidth={2} />
      <circle cx={12} cy={12} r={3} fill="none" stroke="currentColor" strokeWidth={2} />
      <g stroke="currentColor" strokeWidth={1.6}>
        <line x1={12} y1={5} x2={12} y2={9} />
        <line x1={18} y1={15.5} x2={14.5} y2={13.5} />
        <line x1={6} y1={15.5} x2={9.5} y2={13.5} />
      </g>
    </svg>
  );
}

export function SearchIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" width="1em" height="1em" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" aria-hidden="true" {...props}>
      <circle cx={11} cy={11} r={7} />
      <line x1={16.5} y1={16.5} x2={21} y2={21} />
    </svg>
  );
}

export function CartIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" width="1em" height="1em" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <path d="M3 4h2l2.2 11.2a1.5 1.5 0 0 0 1.5 1.2h8.1a1.5 1.5 0 0 0 1.5-1.2L21 8H6" />
      <circle cx={9.5} cy={20} r={1.4} />
      <circle cx={17.5} cy={20} r={1.4} />
    </svg>
  );
}

export function MenuIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" width="1em" height="1em" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" aria-hidden="true" {...props}>
      <line x1={4} y1={7} x2={20} y2={7} />
      <line x1={4} y1={12} x2={20} y2={12} />
      <line x1={4} y1={17} x2={20} y2={17} />
    </svg>
  );
}

export function CloseIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" width="1em" height="1em" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" aria-hidden="true" {...props}>
      <line x1={6} y1={6} x2={18} y2={18} />
      <line x1={18} y1={6} x2={6} y2={18} />
    </svg>
  );
}

export function ChevronDown(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" width="1em" height="1em" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>
      <path d="M6 9l6 6 6-6" />
    </svg>
  );
}
