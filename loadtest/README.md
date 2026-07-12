# Load test

A [k6](https://k6.io) driver that generates believable browse traffic (plus a slice of real checkouts)
against the bike shop. It exists to make two platform showcases visible on a running environment:

- **Autoscaling (ADR-078)** — sustained CPU on the storefront trips its HPA (`minReplicas: 2` →
  `maxReplicas: 6`); if the cluster can't fit the new pods, Karpenter provisions a node.
- **Progressive delivery (ADR-056)** — the storefront's metric-gated canary needs real HTTP
  success-rate data; this traffic feeds the `storefront-canary-gate` AnalysisTemplate.

It runs **outside** the cluster, against the public shop host — no in-cluster load service, so no Kyverno
team-ECR image constraint. The shop's EKS API is private, but the storefront host is public; you only need
network reach to `shop-alpha-dev.preprod.aws.refplat.org`.

## Run

```bash
# defaults: 40 peak VUs, 5m steady, 15% of iterations check out
k6 run loadtest/shop.js

# point at another stage / tune the load
SHOP_URL=https://shop-alpha-dev.preprod.aws.refplat.org VUS=60 DURATION=10m CHECKOUT_RATIO=0.2 \
  k6 run loadtest/shop.js
```

## Watch the platform react

```bash
# storefront replicas climbing (HPA) — over Tailscale to the private EKS API
kubectl --context preprod -n alpha-shop-dev get hpa storefront -w
kubectl --context preprod -n alpha-shop-dev get rollout storefront -w

# Karpenter adding nodes, if scheduling pressure demands it
kubectl --context preprod get nodes -l karpenter.sh/nodepool -w
```
