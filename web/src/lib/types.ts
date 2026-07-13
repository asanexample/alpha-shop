// Mirror of the catalog domain model served by the BFF (see internal/catalog/model.go).
// brand/category on a Product are SLUGS — map to display names via the nav data.

export type Kind =
  | "bikes"
  | "ebikes"
  | "components"
  | "wheels"
  | "tires"
  | "accessories"
  | "apparel";

export interface Category {
  slug: string;
  name: string;
  kind: Kind;
  blurb?: string;
}

export interface Brand {
  slug: string;
  name: string;
}

export interface Product {
  id: string;
  slug: string;
  name: string;
  brand: string; // Brand.slug
  category: string; // Category.slug
  priceCents: number;
  salePriceCents?: number; // > 0 and < priceCents ⇒ on sale
  summary: string;
  description: string;
  image: string;
  sizes?: string[];
  specs?: Record<string, string>;
  inStock: boolean;
  featured?: boolean;
}

export interface NavData {
  categories: Category[];
  brands: Brand[];
}

export interface HomeData {
  featured: Product[];
  categories: Category[];
}

// GET /api/products (and /api/catalog/products) is a page of results — page/perPage are 1-indexed /
// as-served (echoed back so the pager can compute page count from total).
export interface ProductList {
  products: Product[];
  total: number;
  page: number;
  perPage: number;
}

export interface ProductDetail {
  product: Product;
  related: Product[];
}

// The kind → display-label + ordering used by the mega-menu and category grouping.
export const KIND_ORDER: Kind[] = [
  "bikes",
  "ebikes",
  "components",
  "wheels",
  "tires",
  "accessories",
  "apparel",
];

export const KIND_LABEL: Record<Kind, string> = {
  bikes: "Bikes",
  ebikes: "E-Bikes",
  components: "Components",
  wheels: "Wheels",
  tires: "Tires",
  accessories: "Accessories",
  apparel: "Apparel",
};

// ---- Buy path: cart + orders (mirrors internal/cart + internal/orders) ----

// One line in the cart / an order. priceCents is snapshotted at add-time.
export interface CartLine {
  productId: string;
  slug?: string;
  name: string;
  priceCents: number;
  qty: number;
}

export interface Cart {
  sessionId: string;
  items: CartLine[];
  updatedAt: string;
}

// GET /api/cart and the cart mutations return the cart plus derived totals.
export interface CartEnvelope {
  cart: Cart;
  count: number;
  subtotalCents: number;
}

export type OrderStatus = "placed" | "declined";

// Shipping/billing address captured at checkout (internal/orders.Address).
export interface Address {
  name: string;
  line1: string;
  line2?: string;
  city: string;
  state: string;
  zip: string;
  country: string;
}

export type PaymentMethod = "card" | "paypal" | "apple_pay";

export const PAYMENT_METHOD_LABEL: Record<PaymentMethod, string> = {
  card: "Credit / debit card",
  paypal: "PayPal",
  apple_pay: "Apple Pay",
};

// A completed (or declined) checkout. reason is present only when declined.
export interface Order {
  id: string;
  sessionId: string;
  lines: CartLine[];
  totalCents: number;
  status: OrderStatus;
  paymentId: string;
  reason?: string;
  experience?: string; // flagship-flagged checkout variant (standard | express)
  shipping?: string;
  shipmentId?: string; // Bravo Dispatch tracking number (e.g. "BD-10023"), when dispatch was reachable
  address?: Address;
  paymentMethod?: PaymentMethod;
  createdAt: string;
}

// GET /api/orders list item (internal/orders.Summary) — enough for an order-history row.
export interface OrderSummary {
  id: string;
  status: OrderStatus;
  totalCents: number;
  createdAt: string;
}

// ---- Accounts (mirrors cmd/storefront/auth.go's authUser) ----
export interface AuthUser {
  userId: string;
  email: string;
  name: string;
}

// A product carries slugs; true = on sale.
export function isOnSale(p: Product): boolean {
  return typeof p.salePriceCents === "number" && p.salePriceCents > 0 && p.salePriceCents < p.priceCents;
}

// The price a shopper actually pays (sale price when on sale).
export function effectiveCents(p: Product): number {
  return isOnSale(p) ? (p.salePriceCents as number) : p.priceCents;
}
