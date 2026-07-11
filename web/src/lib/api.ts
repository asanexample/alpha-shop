// Same-origin BFF client. All endpoints are GET + JSON (see cmd/storefront/main.go).
import type { HomeData, NavData, ProductDetail, ProductList } from "./types";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

async function getJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
  let res: Response;
  try {
    res = await fetch(path, { signal, headers: { Accept: "application/json" } });
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") throw err;
    throw new ApiError("Can't reach the shop right now. Check your connection and try again.", 0);
  }
  if (!res.ok) {
    // The BFF returns { error } for 404 / 502.
    let detail = "";
    try {
      const body = (await res.json()) as { error?: string };
      detail = body?.error ?? "";
    } catch {
      /* ignore parse errors */
    }
    throw new ApiError(detail || `Request failed (${res.status}).`, res.status);
  }
  return (await res.json()) as T;
}

// Filters accepted by GET /api/products. minPrice/maxPrice are whole DOLLARS.
export interface ProductQuery {
  category?: string;
  brand?: string;
  kind?: string;
  q?: string;
  minPrice?: number;
  maxPrice?: number;
  onSale?: boolean;
  featured?: boolean;
}

export function productSearchParams(q: ProductQuery): URLSearchParams {
  const p = new URLSearchParams();
  if (q.category) p.set("category", q.category);
  if (q.brand) p.set("brand", q.brand);
  if (q.kind) p.set("kind", q.kind);
  if (q.q) p.set("q", q.q);
  if (q.minPrice != null && q.minPrice > 0) p.set("minPrice", String(q.minPrice));
  if (q.maxPrice != null && q.maxPrice > 0) p.set("maxPrice", String(q.maxPrice));
  if (q.onSale) p.set("onSale", "true");
  if (q.featured) p.set("featured", "true");
  return p;
}

export const api = {
  nav: (signal?: AbortSignal) => getJSON<NavData>("/api/nav", signal),
  home: (signal?: AbortSignal) => getJSON<HomeData>("/api/home", signal),
  products: (q: ProductQuery, signal?: AbortSignal) => {
    const qs = productSearchParams(q).toString();
    return getJSON<ProductList>(`/api/products${qs ? `?${qs}` : ""}`, signal);
  },
  product: (idOrSlug: string, signal?: AbortSignal) =>
    getJSON<ProductDetail>(`/api/products/${encodeURIComponent(idOrSlug)}`, signal),
};
