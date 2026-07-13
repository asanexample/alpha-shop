// BFF-backed cart state (TanStack Query). GET /api/cart is the source of truth; add/remove/clear are
// mutations that invalidate it. The provider also owns the toast queue used by the Toaster.
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { api } from "../lib/api";
import type { CartLine } from "../lib/types";

// The snapshot needed to add a product to the cart (qty defaults to 1).
export type CartAddInput = Omit<CartLine, "qty">;

export const CART_QUERY_KEY = ["cart"] as const;

interface Toast {
  id: number;
  message: string;
}

interface CartValue {
  // Cart data (derived from the BFF envelope).
  items: CartLine[];
  count: number;
  subtotalCents: number;
  isLoading: boolean;
  isError: boolean;

  // Mutations.
  addItem: (item: CartAddInput, qty?: number) => void;
  setQty: (productId: string, qty: number) => void;
  removeItem: (productId: string) => void;
  clear: () => void;
  isAdding: boolean;
  isClearing: boolean;
  removingId: string | null;
  updatingId: string | null;

  // Toasts.
  notify: (message: string) => void;
  toasts: Toast[];
  dismiss: (id: number) => void;
}

const CartContext = createContext<CartValue | null>(null);

export function CartProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();

  // ---- Toasts ----
  const [toasts, setToasts] = useState<Toast[]>([]);
  const nextId = useRef(1);
  const timers = useRef<Map<number, ReturnType<typeof setTimeout>>>(new Map());

  const dismiss = useCallback((id: number) => {
    setToasts((t) => t.filter((x) => x.id !== id));
    const timer = timers.current.get(id);
    if (timer) {
      clearTimeout(timer);
      timers.current.delete(id);
    }
  }, []);

  const pushToast = useCallback(
    (message: string) => {
      const id = nextId.current++;
      setToasts((t) => [...t, { id, message }]);
      const timer = setTimeout(() => dismiss(id), 3200);
      timers.current.set(id, timer);
    },
    [dismiss],
  );

  // ---- Cart query ----
  const cartQuery = useQuery({
    queryKey: CART_QUERY_KEY,
    queryFn: ({ signal }) => api.cart.get(signal),
    staleTime: 10_000,
  });

  const invalidateCart = useCallback(
    () => queryClient.invalidateQueries({ queryKey: CART_QUERY_KEY }),
    [queryClient],
  );

  // ---- Mutations ----
  const addMut = useMutation({
    mutationFn: ({ item, qty }: { item: CartAddInput; qty: number }) =>
      api.cart.add({ ...item, qty }),
    onSuccess: (env, vars) => {
      queryClient.setQueryData(CART_QUERY_KEY, env);
      void invalidateCart();
      pushToast(`Added to cart — ${vars.item.name}`);
    },
    onError: () => pushToast("Couldn't add that to your cart. Please try again."),
  });

  const setQtyMut = useMutation({
    mutationFn: ({ productId, qty }: { productId: string; qty: number }) =>
      api.cart.setQty(productId, qty),
    onSuccess: (env) => {
      queryClient.setQueryData(CART_QUERY_KEY, env);
      void invalidateCart();
    },
    onError: () => pushToast("Couldn't update that quantity. Please try again."),
  });

  const removeMut = useMutation({
    mutationFn: (productId: string) => api.cart.remove(productId),
    onSuccess: (env) => {
      queryClient.setQueryData(CART_QUERY_KEY, env);
      void invalidateCart();
    },
    onError: () => pushToast("Couldn't remove that item. Please try again."),
  });

  const clearMut = useMutation({
    mutationFn: () => api.cart.clear(),
    onSuccess: () => invalidateCart(),
    onError: () => pushToast("Couldn't empty your cart. Please try again."),
  });

  const addItem = useCallback(
    (item: CartAddInput, qty = 1) => addMut.mutate({ item, qty }),
    [addMut],
  );
  const setQty = useCallback(
    (productId: string, qty: number) => setQtyMut.mutate({ productId, qty }),
    [setQtyMut],
  );
  const removeItem = useCallback((productId: string) => removeMut.mutate(productId), [removeMut]);
  const clear = useCallback(() => clearMut.mutate(), [clearMut]);

  const env = cartQuery.data;
  const removingId = removeMut.isPending ? (removeMut.variables ?? null) : null;
  const updatingId = setQtyMut.isPending ? (setQtyMut.variables?.productId ?? null) : null;

  const value = useMemo<CartValue>(
    () => ({
      items: env?.cart.items ?? [],
      count: env?.count ?? 0,
      subtotalCents: env?.subtotalCents ?? 0,
      isLoading: cartQuery.isLoading,
      isError: cartQuery.isError,
      addItem,
      setQty,
      removeItem,
      clear,
      isAdding: addMut.isPending,
      isClearing: clearMut.isPending,
      removingId,
      updatingId,
      notify: pushToast,
      toasts,
      dismiss,
    }),
    [
      env,
      cartQuery.isLoading,
      cartQuery.isError,
      addItem,
      setQty,
      removeItem,
      clear,
      addMut.isPending,
      clearMut.isPending,
      removingId,
      updatingId,
      pushToast,
      toasts,
      dismiss,
    ],
  );

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart(): CartValue {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error("useCart must be used within CartProvider");
  return ctx;
}
