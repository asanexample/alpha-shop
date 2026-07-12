// A stable per-visitor cart session id, minted once and persisted in localStorage. The buy-path BFF
// endpoints key a cart/order off this via the X-Shop-Session header (see cmd/storefront/buypath.go).
const STORAGE_KEY = "alpha-shop.sid";

// In-memory fallback for environments where localStorage is unavailable (private mode / SSR).
let memoId: string | null = null;

function mint(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  // Last-resort fallback — good enough to key a demo cart.
  return `sid-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

/** Returns this browser's cart session id, creating and persisting one on first use. */
export function getSessionId(): string {
  if (memoId) return memoId;
  try {
    let sid = localStorage.getItem(STORAGE_KEY);
    if (!sid) {
      sid = mint();
      localStorage.setItem(STORAGE_KEY, sid);
    }
    memoId = sid;
    return sid;
  } catch {
    // localStorage blocked — hold the id in memory for the life of the tab.
    memoId ??= mint();
    return memoId;
  }
}
