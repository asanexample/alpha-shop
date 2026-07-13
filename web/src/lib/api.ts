// Same-origin BFF client. Browse endpoints are plain GET + JSON; the buy-path (cart/checkout/orders)
// endpoints additionally carry the X-Shop-Session header (see cmd/storefront/buypath.go).
import { getSessionId } from "./session";
import type {
  Address,
  AuthUser,
  CartEnvelope,
  CartLine,
  HomeData,
  NavData,
  Order,
  OrderSummary,
  PaymentMethod,
  ProductDetail,
  ProductList,
} from "./types";

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

// Session-aware request for the buy-path endpoints. Attaches X-Shop-Session, serialises a JSON body
// when given, and tolerates a 204 (empty) response.
async function sessionJSON<T>(
  path: string,
  init: { method?: string; body?: unknown; signal?: AbortSignal } = {},
): Promise<T> {
  const headers: Record<string, string> = {
    Accept: "application/json",
    "X-Shop-Session": getSessionId(),
  };
  if (init.body !== undefined) headers["Content-Type"] = "application/json";

  let res: Response;
  try {
    res = await fetch(path, {
      method: init.method ?? "GET",
      headers,
      body: init.body !== undefined ? JSON.stringify(init.body) : undefined,
      signal: init.signal,
    });
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") throw err;
    throw new ApiError("Can't reach the shop right now. Check your connection and try again.", 0);
  }
  if (!res.ok) {
    // The BFF returns { error } for 400 / 404 / 502.
    let detail = "";
    try {
      const body = (await res.json()) as { error?: string };
      detail = body?.error ?? "";
    } catch {
      /* ignore parse errors */
    }
    throw new ApiError(detail || `Request failed (${res.status}).`, res.status);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

// Filters accepted by GET /api/products. minPrice/maxPrice are whole DOLLARS. page/perPage are
// 1-indexed; omitted ⇒ the BFF/catalog defaults (page 1, 24 per page).
export interface ProductQuery {
  category?: string;
  brand?: string;
  kind?: string;
  q?: string;
  minPrice?: number;
  maxPrice?: number;
  onSale?: boolean;
  featured?: boolean;
  page?: number;
  perPage?: number;
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
  if (q.page != null && q.page > 1) p.set("page", String(q.page));
  if (q.perPage != null && q.perPage > 0) p.set("perPage", String(q.perPage));
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

  // ---- Buy path (session header) ----
  cart: {
    get: (signal?: AbortSignal) => sessionJSON<CartEnvelope>("/api/cart", { signal }),
    add: (line: CartLine) =>
      sessionJSON<CartEnvelope>("/api/cart/items", { method: "POST", body: line }),
    setQty: (productId: string, qty: number) =>
      sessionJSON<CartEnvelope>(`/api/cart/items/${encodeURIComponent(productId)}`, {
        method: "PATCH",
        body: { qty },
      }),
    remove: (productId: string) =>
      sessionJSON<CartEnvelope>(`/api/cart/items/${encodeURIComponent(productId)}`, {
        method: "DELETE",
      }),
    clear: () => sessionJSON<void>("/api/cart", { method: "DELETE" }),
  },
  // Checkout requires a signed-in session (the BFF resolves the account from the auth cookie) and
  // returns HTTP 200 for both "placed" and "declined"; a 400 { error } means the cart is empty, a 401
  // means "sign in first".
  checkout: (input: { card: string; address: Address; paymentMethod: PaymentMethod }) =>
    sessionJSON<Order>("/api/checkout", { method: "POST", body: input }),
  order: (id: string, signal?: AbortSignal) =>
    sessionJSON<Order>(`/api/orders/${encodeURIComponent(id)}`, { signal }),
  orders: (signal?: AbortSignal) =>
    sessionJSON<{ orders: OrderSummary[] }>("/api/orders", { signal }),

  // ---- Auth (cookie session — accounts is the sole authority on identity) ----
  auth: {
    signup: (input: { email: string; password: string; name: string }) =>
      sessionJSON<AuthUser>("/api/auth/signup", { method: "POST", body: input }),
    login: (input: { email: string; password: string }) =>
      sessionJSON<AuthUser>("/api/auth/login", { method: "POST", body: input }),
    logout: () => sessionJSON<void>("/api/auth/logout", { method: "POST" }),
    me: (signal?: AbortSignal) => sessionJSON<AuthUser>("/api/auth/me", { signal }),
  },
};
