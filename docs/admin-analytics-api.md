# Admin Analytics API

> **Status:** Planned — not yet implemented in `rahil-gallery-server`.  
> **Client spec:** [kpi-analytics.md](../../rahil-gallery-client/docs/kpi-analytics.md)  
> **Auth:** Staff JWT + **Admin** role required on all endpoints.

---

## Overview

Five read-only endpoints power the admin analytics dashboards. Each accepts a `period` query parameter and returns aggregates computed server-side from orders, customers, catalog, inventory, and (future) session/event data.

| Endpoint | Dashboard tab |
|----------|---------------|
| `GET /api/v1/admin/analytics/overview` | Executive |
| `GET /api/v1/admin/analytics/marketing` | Marketing |
| `GET /api/v1/admin/analytics/product` | Product |
| `GET /api/v1/admin/analytics/customer` | Customer |
| `GET /api/v1/admin/analytics/funnel` | Funnel |

---

## Common query parameters

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `period` | `7d` \| `30d` \| `90d` | `30d` | Rolling window ending at request time (UTC) |

### Period semantics

- **Window:** `[now − period_days, now)` in UTC.
- **Comparison:** Prior equal-length window immediately before the current window.
- **Change fields:** Percent or percentage-point delta vs comparison window.

### Response envelope

```json
{
  "success": true,
  "data": { },
  "meta": {
    "period": "30d",
    "periodStart": "2026-05-11T00:00:00Z",
    "periodEnd": "2026-06-10T00:00:00Z",
    "comparisonStart": "2026-04-11T00:00:00Z",
    "comparisonEnd": "2026-05-11T00:00:00Z",
    "generatedAt": "2026-06-10T08:00:00Z"
  }
}
```

Monetary values are **integers in IRR (Riam)** unless noted. Rates are **floats** (e.g. `2.8` = 2.8%).

---

## GET `/api/v1/admin/analytics/overview`

Executive summary — core business KPIs and top products.

### Response `data`

```json
{
  "kpis": {
    "netRevenue": 1200000000,
    "orders": 342,
    "conversionRate": 2.8,
    "aov": 3500000,
    "cac": 480000,
    "ltv": 8200000,
    "ltvCacRatio": 17.1,
    "grossMarginPercent": 62.4,
    "returnRate": 3.2,
    "roas": 4.6
  },
  "comparison": {
    "netRevenueChangePercent": 8.4,
    "ordersChangePercent": 12.0,
    "conversionRateChangePp": 0.3,
    "aovChangePercent": 4.1,
    "cacChangePercent": -6.0,
    "ltvChangePercent": 2.1,
    "grossMarginChangePp": 0.8,
    "returnRateChangePp": -0.4,
    "roasChange": 0.5
  },
  "topProducts": [
    {
      "sku": "RNG-SOL-1CT",
      "name": "Solitaire ring · 1 ct diamond",
      "revenue": 185000000,
      "sharePercent": 18.2
    }
  ]
}
```

### KPI definitions

| Field | Formula |
|-------|---------|
| `netRevenue` | Gross revenue − discounts − returns − cancelled value |
| `orders` | Count of non-cancelled, paid orders |
| `conversionRate` | `orders / sessions × 100` |
| `aov` | `netRevenue / orders` |
| `cac` | `marketing_spend / new_customers` |
| `ltv` | Cohort-based average customer lifetime revenue |
| `ltvCacRatio` | `ltv / cac` |
| `grossMarginPercent` | `(revenue − COGS) / revenue × 100` |
| `returnRate` | `returned_orders / delivered_orders × 100` |
| `roas` | `attributed_ad_revenue / ad_spend` |

---

## GET `/api/v1/admin/analytics/marketing`

Channel mix, CAC, influencer metrics, and campaign table.

### Response `data`

```json
{
  "kpis": {
    "blendedRoas": 4.6,
    "totalAdSpend": 142000000,
    "emailRevenueShare": 14.2,
    "organicRevenueShare": 38.5
  },
  "comparison": {
    "totalAdSpendChangePercent": 5.0,
    "emailRevenueShareChangePp": 1.1
  },
  "channelRevenueMix": [
    { "channel": "organic_search", "sharePercent": 38.5, "revenue": 392000000 }
  ],
  "cacByChannel": [
    { "channel": "paid_search", "cac": 520000, "note": "Highest volume" }
  ],
  "influencerMetrics": {
    "conversionRate": 8.4,
    "avgOrdersPerCampaign": 24,
    "revenuePerPost": 11200000,
    "roas": 3.8
  },
  "campaigns": [
    {
      "name": "Spring bridal · Meta",
      "channel": "social",
      "spend": 28000000,
      "revenue": 145000000,
      "roas": 5.2,
      "ctr": 2.4,
      "conversionRate": 3.1
    }
  ]
}
```

### Channel taxonomy

`organic_search` · `paid_search` · `social_paid` · `email` · `influencer` · `direct_other`

---

## GET `/api/v1/admin/analytics/product`

SKU performance, conversion rates, inventory health.

### Response `data`

```json
{
  "kpis": {
    "viewToCartRate": 12.4,
    "cartToPurchaseRate": 22.6,
    "inventoryTurnover": 4.2,
    "stockoutRate": 2.1
  },
  "comparison": {
    "viewToCartChangePp": 0.6,
    "cartToPurchaseChangePp": -0.2
  },
  "skuRevenueRanking": [
    { "sku": "RNG-SOL-1CT", "name": "Solitaire ring · 1 ct", "sharePercent": 18.2, "revenue": 185000000 }
  ],
  "skuMarginRanking": [
    { "sku": "RNG-CUSTOM", "name": "Custom engagement", "marginPercent": 68.0 }
  ],
  "inventoryStatus": [
    { "status": "in_stock", "skuCount": 142 },
    { "status": "made_to_order", "skuCount": 48 },
    { "status": "out_of_stock", "skuCount": 4 },
    { "status": "low_stock", "skuCount": 11 }
  ]
}
```

---

## GET `/api/v1/admin/analytics/customer`

Lifecycle metrics, first-purchase mix, cohort table.

### Response `data`

```json
{
  "kpis": {
    "ltv": 8200000,
    "repeatPurchaseRate": 28.4,
    "avgDaysBetweenPurchases": 142,
    "wishlistRate": 18.6,
    "cartAbandonmentRate": 71.2
  },
  "comparison": {
    "ltvChangePercent": 2.1,
    "repeatPurchaseChangePp": 1.8,
    "cartAbandonmentChangePp": -1.4
  },
  "firstPurchaseCategories": [
    { "category": "rings", "sharePercent": 42.1, "note": "Engagement-led" }
  ],
  "cohorts": [
    {
      "month": "2026-01",
      "customers": 84,
      "repeatRate": 22.0,
      "avgLtv": 6800000
    }
  ]
}
```

---

## GET `/api/v1/admin/analytics/funnel`

Ecommerce funnel, device split, checkout friction, payment failures.

### Response `data`

```json
{
  "kpis": {
    "productViewRate": 68.4,
    "cartConversionRate": 12.4,
    "overallConversionRate": 2.8,
    "checkoutDropOffRate": 38.2,
    "paymentFailureRate": 4.8
  },
  "funnel": [
    { "stage": "sessions", "count": 12240, "rateFromSession": 100 },
    { "stage": "product_view", "count": 8372, "rateFromPrevious": 68.4, "rateFromSession": 68.4, "dropOff": 31.6 },
    { "stage": "add_to_cart", "count": 1038, "rateFromPrevious": 12.4, "rateFromSession": 8.5, "dropOff": 87.6 },
    { "stage": "checkout_started", "count": 642, "rateFromPrevious": 61.8, "rateFromSession": 5.2, "dropOff": 38.2 },
    { "stage": "purchase", "count": 342, "rateFromPrevious": 53.3, "rateFromSession": 2.8, "dropOff": 46.7 }
  ],
  "deviceBreakdown": [
    { "device": "mobile", "sharePercent": 62.4, "conversionRate": 2.4 }
  ],
  "checkoutDropOffByStep": [
    { "step": "address", "sharePercent": 12.4 }
  ],
  "paymentFailureBreakdown": [
    { "reason": "insufficient_funds", "sharePercent": 42.0 }
  ]
}
```

---

## Data sources (planned)

| Source | KPIs powered |
|--------|--------------|
| `orders`, `order_items`, `payments` | Revenue, AOV, orders, funnel purchase step |
| `returns` | Return rate, net revenue adjustment |
| `products`, `product_variants`, `inventory_items` | SKU rankings, stockout, turnover |
| `customer_profiles`, `users` | LTV, repeat rate, cohorts |
| `marketing_spend`, `campaigns` ⚠️ | ROAS, CAC, channel mix |
| `analytics_events` ⚠️ | Sessions, PDP views, cart, checkout, device |

---

## Implementation notes

1. **Materialized views** recommended for executive KPIs (refresh nightly or hourly).
2. **Session table** required before funnel and conversion metrics are authoritative.
3. **COGS** per SKU needed for margin KPIs (see AN-03 in client open decisions).
4. **Caching:** `Cache-Control: private, max-age=300` acceptable for v1 batch aggregates.
5. **403** when authenticated user lacks `admin` role.

### Suggested package layout

```
internal/
├── domain/analytics/          # KPI types, period parsing
├── application/analytics/   # Aggregation use cases
├── infrastructure/
│   └── persistence/postgres/analytics/
└── interfaces/http/handler/admin_analytics.go
```

---

## Client integration checklist

- [ ] Create `lib/api/admin/analytics.ts` with typed fetchers per endpoint
- [ ] Map API responses → `AnalyticsSnapshot` shape in `analytics-data.ts`
- [ ] Replace mock scaling with live `comparison` deltas from API
- [ ] Add loading/error UI in analytics views
- [ ] Enforce Admin role in `AdminAuthGate` for `/admin/analytics/*`

---

*Cross-reference: [admin-customers-api.md](./admin-customers-api.md) for CRM filters; [technical-workflow.md](./technical-workflow.md) for dev setup.*
