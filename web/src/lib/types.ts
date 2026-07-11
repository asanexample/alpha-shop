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

export interface ProductList {
  products: Product[];
  count: number;
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

// A product carries slugs; true = on sale.
export function isOnSale(p: Product): boolean {
  return typeof p.salePriceCents === "number" && p.salePriceCents > 0 && p.salePriceCents < p.priceCents;
}

// The price a shopper actually pays (sale price when on sale).
export function effectiveCents(p: Product): number {
  return isOnSale(p) ? (p.salePriceCents as number) : p.priceCents;
}
