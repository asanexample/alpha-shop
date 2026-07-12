// k6 load driver for the bike shop — exercises the browse path continuously with a slice of real checkouts.
//
// Purpose (ADR-056 / ADR-078 showcases): generate steady, believable traffic against the storefront so that
//   • the storefront HPA scales on CPU (and Karpenter adds nodes if the cluster can't fit the pods), and
//   • a storefront canary's metric gate has real HTTP success-rate data to analyse.
//
// Runs OUTSIDE the cluster (over Tailscale, against the public shop host) — no in-cluster load service, so no
// Kyverno team-ECR image constraint. See loadtest/README.md.
//
//   SHOP_URL=https://shop-alpha-dev.preprod.aws.refplat.org k6 run loadtest/shop.js
//
// Tunables (env): VUS (peak virtual users, default 40), DURATION (steady phase, default 5m),
//                 CHECKOUT_RATIO (fraction of iterations that buy, default 0.15).

import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";

const BASE = __ENV.SHOP_URL || "https://shop-alpha-dev.preprod.aws.refplat.org";
const PEAK = parseInt(__ENV.VUS || "40", 10);
const DURATION = __ENV.DURATION || "5m";
const CHECKOUT_RATIO = parseFloat(__ENV.CHECKOUT_RATIO || "0.15");

const checkouts = new Counter("shop_checkouts");

export const options = {
  // Ramp up so the HPA/Karpenter reaction is visible, hold, then ease off.
  stages: [
    { duration: "1m", target: PEAK },
    { duration: DURATION, target: PEAK },
    { duration: "1m", target: 0 },
  ],
  thresholds: {
    http_req_failed: ["rate<0.05"],
    http_req_duration: ["p(95)<2000"],
  },
};

// setup runs once: pull the live catalog so the script never hardcodes ids/prices (they'd drift from the seed).
export function setup() {
  const res = http.get(`${BASE}/api/products`);
  check(res, { "catalog reachable": (r) => r.status === 200 });
  const products = (res.json("products") || []).map((p) => ({
    id: p.id,
    slug: p.slug,
    name: p.name,
    priceCents: p.priceCents,
  }));
  if (products.length === 0) throw new Error("no products returned from /api/products");
  return { products };
}

function pick(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

export default function (data) {
  // Browse — the bulk of real shop traffic. These GETs drive the storefront's CPU + request metrics.
  http.get(`${BASE}/`);
  http.get(`${BASE}/api/home`);
  http.get(`${BASE}/api/products`);
  const p = pick(data.products);
  http.get(`${BASE}/api/products?category=${encodeURIComponent(p.category || "")}`);

  // A slice of visitors actually buy — add to cart + checkout (storefront → orders → payment east-west).
  if (Math.random() < CHECKOUT_RATIO) {
    const sid = `loadtest-${__VU}-${__ITER}`;
    const headers = { "X-Shop-Session": sid, "Content-Type": "application/json" };
    http.post(
      `${BASE}/api/cart/items`,
      JSON.stringify({ productId: p.id, slug: p.slug, name: p.name, priceCents: p.priceCents, qty: 1 }),
      { headers },
    );
    const co = http.post(`${BASE}/api/checkout`, JSON.stringify({ card: "4242424242424242" }), { headers });
    check(co, { "checkout ok": (r) => r.status === 200 });
    checkouts.add(1);
  }

  sleep(Math.random() * 2 + 0.5); // 0.5–2.5s think time
}
