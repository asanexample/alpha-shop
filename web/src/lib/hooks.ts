// TanStack Query hooks for the BFF, plus a nav-lookup helper that turns the slug-based product
// fields into display names (brand/category on a Product are slugs).
import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import { api, type ProductQuery } from "./api";
import type { Category, Kind } from "./types";

export function useNav() {
  return useQuery({
    queryKey: ["nav"],
    queryFn: ({ signal }) => api.nav(signal),
    staleTime: Infinity, // nav is effectively static for a session
  });
}

export function useHome() {
  return useQuery({
    queryKey: ["home"],
    queryFn: ({ signal }) => api.home(signal),
  });
}

export function useProducts(query: ProductQuery) {
  return useQuery({
    queryKey: ["products", query],
    queryFn: ({ signal }) => api.products(query, signal),
    placeholderData: (prev) => prev, // keep last results visible while re-filtering
  });
}

export function useProduct(idOrSlug: string | undefined) {
  return useQuery({
    queryKey: ["product", idOrSlug],
    queryFn: ({ signal }) => api.product(idOrSlug as string, signal),
    enabled: !!idOrSlug,
    retry: (count, err) => {
      // Don't retry a genuine 404.
      const status = (err as { status?: number })?.status;
      return status === 404 ? false : count < 2;
    },
  });
}

export interface NavLookups {
  categoryName: (slug: string) => string;
  categoryKind: (slug: string) => Kind | undefined;
  category: (slug: string) => Category | undefined;
  brandName: (slug: string) => string;
}

/** Slug → display-name lookups derived from nav data (cached; safe to call before nav loads). */
export function useNavLookups(): NavLookups {
  const { data } = useNav();
  return useMemo(() => {
    const catBySlug = new Map((data?.categories ?? []).map((c) => [c.slug, c]));
    const brandBySlug = new Map((data?.brands ?? []).map((b) => [b.slug, b]));
    return {
      category: (slug) => catBySlug.get(slug),
      categoryName: (slug) => catBySlug.get(slug)?.name ?? slug,
      categoryKind: (slug) => catBySlug.get(slug)?.kind,
      brandName: (slug) => brandBySlug.get(slug)?.name ?? slug,
    };
  }, [data]);
}
