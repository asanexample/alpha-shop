// Local-only cart (P2 stub): clicking "Add to cart" increments a client-side count and shows a
// toast. No server call yet — the BFF cart/orders endpoints come next.
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";

interface Toast {
  id: number;
  message: string;
}

interface CartValue {
  count: number;
  add: (label: string, qty?: number) => void;
  notify: (message: string) => void;
  toasts: Toast[];
  dismiss: (id: number) => void;
}

const CartContext = createContext<CartValue | null>(null);

export function CartProvider({ children }: { children: ReactNode }) {
  const [count, setCount] = useState(0);
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

  const add = useCallback(
    (label: string, qty = 1) => {
      setCount((c) => c + qty);
      pushToast(`Added to cart — ${label}`);
    },
    [pushToast],
  );

  const value = useMemo<CartValue>(
    () => ({ count, add, notify: pushToast, toasts, dismiss }),
    [count, add, pushToast, toasts, dismiss],
  );

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>;
}

export function useCart(): CartValue {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error("useCart must be used within CartProvider");
  return ctx;
}
