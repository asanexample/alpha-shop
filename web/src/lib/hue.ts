// Deterministic colour derivation for the blueprint thumbnails. The same product id always
// yields the same hue, so the catalog reads as one cohesive set of workshop drawings.

function hashString(s: string): number {
  // FNV-1a — stable, well-distributed, tiny.
  let h = 0x811c9dc5;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 0x01000193);
  }
  return h >>> 0;
}

export interface Blueprint {
  hue: number; // 0..359
  paper: string; // panel background
  grid: string; // graph-paper line colour
  gridBold: string; // major grid line
  ink: string; // line-icon stroke
}

/**
 * Derive a cohesive blueprint palette from a product id. Every panel is a muted, low-saturation
 * tint (a drafting-paper wash) so the whole grid stays disciplined; the yellow accent is layered
 * on separately by the Thumb component.
 */
export function blueprintFor(id: string): Blueprint {
  const hue = hashString(id) % 360;
  return {
    hue,
    paper: `hsl(${hue} 22% 96%)`,
    grid: `hsl(${hue} 20% 84%)`,
    gridBold: `hsl(${hue} 22% 74%)`,
    ink: `hsl(${hue} 28% 22%)`,
  };
}
