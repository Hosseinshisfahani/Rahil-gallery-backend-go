# Admin Customers API

CRM endpoints for managing storefront customers. Used by the Rehil Gallery admin UI (`/admin/customers`).

**Base path:** `/api/v1/admin/customers`  
**Auth:** Bearer JWT (`admin` or `staff` role required)  
**Migrations:** `000002_customer_crm`, `000004_customer_list_perf` (denormalized list stats + search indexes)

---

## Table of contents

1. [Overview](#1-overview)
2. [Authentication](#2-authentication)
3. [Response format](#3-response-format)
4. [Endpoints](#4-endpoints)
5. [Data types](#5-data-types)
6. [Segments & status](#6-segments--status)
7. [Error codes](#7-error-codes)
8. [Examples](#8-examples)
9. [Admin UI integration](#9-admin-ui-integration)
10. [Not yet implemented](#10-not-yet-implemented)

---

## 1. Overview

| Item | Value |
|------|--------|
| Audience | Internal staff (`admin`, `staff` JWT) |
| Customer auth | OTP via phone (storefront); admin-created accounts have **no password** until the customer sets one |
| Customer role | All endpoints operate on users with role `customer` |
| Commerce | Orders, cart, wishlist, and LTV live in Django. This API does not return them. |
| IDs | UUID strings (e.g. `550e8400-e29b-41d4-a716-446655440000`) |

### Endpoint summary

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/customers` | Paginated list with search and filters |
| POST | `/api/v1/admin/customers` | Create customer (quick add or history import) |
| GET | `/api/v1/admin/customers/:id` | Full customer profile |
| PATCH | `/api/v1/admin/customers/:id` | Update profile metadata |
| DELETE | `/api/v1/admin/customers/:id` | Soft-delete account |
| POST | `/api/v1/admin/customers/:id/block` | Block customer |
| POST | `/api/v1/admin/customers/:id/unblock` | Unblock customer |
| POST | `/api/v1/admin/customers/:id/vip` | Toggle VIP status |
| POST | `/api/v1/admin/customers/:id/tags` | Toggle operational tag |
| POST | `/api/v1/admin/customers/:id/notes` | Add internal CRM note |
| GET | `/api/v1/admin/customers/segments` | Built-in segment counts + saved segment shortcuts |
| GET | `/api/v1/admin/customers/saved-views` | List saved filters/segments (`?type=filter\|segment`) |
| POST | `/api/v1/admin/customers/saved-views` | Save a filter or segment shortcut |
| GET | `/api/v1/admin/customers/saved-views/:viewId` | Get saved view |
| PATCH | `/api/v1/admin/customers/saved-views/:viewId` | Update saved view (owner only) |
| DELETE | `/api/v1/admin/customers/saved-views/:viewId` | Delete saved view (owner only) |

---

## 2. Authentication

Every request must include a valid staff access token:

```http
Authorization: Bearer <access_token>
```

Obtain a token via `POST /api/v1/auth/login` with an `admin` or `staff` account.

| Role | Access |
|------|--------|
| `admin` | All customer endpoints |
| `staff` | All customer endpoints (fine-grained roles planned) |
| `customer` | **403 Forbidden** |

Missing or invalid tokens return **401** with:

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "authentication required"
  }
}
```

---

## 3. Response format

Admin customer endpoints use a **flat JSON** envelope — not the `{ "success": true, "data": ... }` wrapper used by `/api/v1/auth/*`.

### List response

```json
{
  "data": [ /* CustomerSummary[] */ ],
  "meta": {
    "page": 1,
    "perPage": 10,
    "total": 42,
    "totalPages": 5
  }
}
```

### Detail / mutation response

Single `CustomerDetail` object (no wrapper).

### Delete response

```json
{ "success": true }
```

### Error response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "human-readable description"
  }
}
```

### Date formats

| Field type | Format | Example |
|------------|--------|---------|
| CRM dates (`createdAt`, birthday, visit dates) | `YYYY-MM-DD` | `"2024-03-12"` |

---

## 4. Endpoints

### GET `/api/v1/admin/customers`

Paginated customer list with **quick search** or **advanced filters** (mutually exclusive).

**Quick search** (search box) — use `q` only. Routes by input shape (no `OR` scan):

| Input | Fields searched |
|-------|-------------------|
| Text (`ali`) | Name only (trigram index) |
| 4+ digits (`98900`) | Phone prefix only (`phone_digits` column) |
| Mixed (`ali98900`) | Name + phone via indexed `UNION` |

**Advanced filters** (filter panel) — use structured params below; when any advanced param is sent, `q` is **ignored**.

| Parameter | Mode | Description |
|-----------|------|-------------|
| `q` | Quick | Name or phone (partial / phone prefix) |
| `page` | Both | Page number (default `1`) |
| `perPage` | Both | Page size (default `10`) |
| `includeTotal` | Both | `false` skips `COUNT(*)`; response uses `meta.hasMore` instead of `total` |

**Advanced-only parameters**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Exact customer ID |
| `email` | string | Partial email match (case-insensitive) |
| `ageRange` | string | Imported CRM age bucket: `1-7` · `7-14` · `14-21` · `21-40` · `40+` |
| `gender` | string | Imported CRM gender: `male` · `female` · `other` |
| `customerTypes` | string | Comma-separated imported customer types |
| `purchaseTypes` | string | Comma-separated purchased categories (overlap match) |
| `firstVisitFrom` | date | First visit on or after |
| `firstVisitTo` | date | First visit on or before |
| `birthdayFrom` | date | Birthday on or after |
| `birthdayTo` | date | Birthday on or before |
| `marriageFrom` | date | Marriage date on or after |
| `marriageTo` | date | Marriage date on or before |

**Response:** `200` — paginated `CustomerSummary` list (see [§5.1](#51-customersummary)).

---

### POST `/api/v1/admin/customers`

Create an admin-provisioned customer account. The account is created with **no password** (`password_hash` is null); the customer signs in via OTP and can set a password later.

#### Quick add

`importMode: "quick"` (default when omitted with `fullName` + `phone`).

| Field | Required | Description |
|-------|:--------:|-------------|
| `importMode` | | `"quick"` |
| `fullName` | ✓ | Display name (split into first / last) |
| `phone` | ✓ | Primary identifier; must be unique |
| `email` | | Optional; must be unique if provided |
| `locale` | | `"fa"` (default) or `"en"` |
| `defaultRingSize` | | Ring size string |
| `isVip` | | Default `false` |
| `tags` | | Operational tag array |

```json
{
  "importMode": "quick",
  "fullName": "Sara Mohammadi",
  "phone": "+989121234567",
  "email": "sara@example.com",
  "locale": "fa",
  "defaultRingSize": "14",
  "isVip": false
}
```

#### History included

`importMode: "history_included"` — full CRM import for gallery walk-ins and legacy migration.

| Field | Required | Description |
|-------|:--------:|-------------|
| `importMode` | ✓ | `"history_included"` |
| `importProfile` | ✓ | See [§5.4 ImportProfile](#54-importprofile) |

```json
{
  "importMode": "history_included",
  "importProfile": {
    "firstName": "Sara",
    "lastName": "Mohammadi",
    "job": "Architect",
    "phone": "+989121234567",
    "email": "sara@example.com",
    "address": "Tehran, Vanak",
    "birthday": "1990-05-12",
    "marriageDate": "2018-03-20",
    "importantDate": "2024-11-01",
    "firstVisitDate": "2024-10-15",
    "customerType": "vip",
    "customerAgeRange": "21-40",
    "purchasedCategories": ["gold_and_gemstones", "silver_and_stones"],
    "description": "Visited for bridal consultation",
    "signature": "S. Mohammadi"
  }
}
```

Selecting `customerType: "vip"` automatically sets `isVip: true`.

**Response:** `201` — full `CustomerDetail`.

**Errors:** `422` validation · `409` duplicate phone or email

---

### GET `/api/v1/admin/customers/:id`

Full CRM profile (contact fields, import profile, and signature). Orders and wishlist are not included.

**Path:** `id` — customer UUID

**Response:** `200` — `CustomerDetail` (see [§5.2](#52-customerdetail))

**Errors:** `404` customer not found or not a `customer` role user

---

### PATCH `/api/v1/admin/customers/:id`

Update customer metadata. All fields are optional; only provided fields are updated. Logged in audit trail as `profile_edit`.

```json
{
  "fullName": "Sara Mohammadi",
  "phone": "+989121234567",
  "email": "sara@example.com",
  "locale": "fa",
  "defaultRingSize": "14",
  "isVip": true,
  "tags": ["VIP", "High spender"],
  "importProfile": { /* same shape as create */ }
}
```

Pass `"email": ""` to clear email.

**Response:** `200` — updated `CustomerDetail`

---

### DELETE `/api/v1/admin/customers/:id`

Soft-delete the customer account (`users.deleted_at` set). Does not purge order history.

**Response:** `200`

```json
{ "success": true }
```

---

### POST `/api/v1/admin/customers/:id/block`

Block a customer. Sets `users.status` to `banned` (exposed as `status: "blocked"` in API).

```json
{
  "reason": "fraud_suspicion",
  "note": "Optional internal note"
}
```

**Reason codes:** `fraud_suspicion` · `payment_issues` · `return_abuse` · `system_misuse`

**Response:** `200` — updated `CustomerDetail` with `blockReason` and optional `blockNote`

Logged in audit trail as `block`.

---

### POST `/api/v1/admin/customers/:id/unblock`

Restore a blocked customer to active status. Requires a justification note.

```json
{
  "justification": "Issue resolved after manual review"
}
```

**Response:** `200` — updated `CustomerDetail`

Logged in audit trail as `unblock`.

---

### POST `/api/v1/admin/customers/:id/vip`

Toggle VIP status. Assigns `vipSource: "manual"` when enabling.

**Request body:** empty or `{}`

**Response:** `200` — updated `CustomerDetail`

Logged as `vip_assign` or `vip_remove`.

---

### POST `/api/v1/admin/customers/:id/tags`

Toggle an operational tag on or off.

```json
{
  "tag": "VIP"
}
```

**Allowed tags:** `VIP` · `Bridal customer` · `High spender` · `At-risk` · `Influencer lead`

**Response:** `200` — updated `CustomerDetail`

Logged as `tag_add` or `tag_remove`.

---

### POST `/api/v1/admin/customers/:id/notes`

Add an internal CRM note (staff-only, visible on profile Notes tab).

```json
{
  "body": "Customer prefers white gold settings."
}
```

**Response:** `200` — updated `CustomerDetail` (includes new note)

Logged as `note_add`.

---

## 5. Data types

### 5.1 CustomerSummary

Returned in list responses.

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | UUID |
| `fullName` | string | First + last name |
| `phone` | string | Primary identifier |
| `customerType` | string | CRM customer type |
| `purchasedCategories` | string[] | CRM product categories from the import profile |
| `customerAgeRange` | string? | Imported age bucket |
| `gender` | string? | Imported gender |
| `createdAt` | string | Record creation date (`YYYY-MM-DD`) |
| `href` | string | Admin UI path, e.g. `/admin/customers/{id}` |

### 5.2 CustomerDetail

Extends `CustomerSummary` with the CRM profile. It does not include orders, wishlist, LTV, or cart metrics.

| Field | Type | Description |
|-------|------|-------------|
| `email` | string? | Optional email |
| `job` | string? | Occupation |
| `address` | string? | Address |
| `melliCode` | string? | National ID |
| `postalCode` | string? | Postal code |
| `birthday` | string? | `YYYY-MM-DD` |
| `marriageDate` | string? | `YYYY-MM-DD` |
| `importantDate` | string? | `YYYY-MM-DD` |
| `firstVisitDate` | string? | `YYYY-MM-DD` |
| `description` | string? | Profile note |
| `marketerNote` | string? | Staff note |
| `signatureUrl` | string? | Public signature image path |
| `importProfile` | object | Same CRM fields used by create/update |

### 5.3 Enums

**`customerType`** (history import): `foreign_and_tour_guidance` · `vip` · `public` · `colleagues` · `family_and_friends`

**`customerAgeRange`:** `1-7` · `7-14` · `14-21` · `21-40` · `40+`

**`purchasedCategories`:** `gold_and_stones` · `silver_and_stones` · `stones_and_roughs` · `gold_and_gemstones` · `silver_and_gemstones` · `gemstones_and_special_roughs`

### 5.4 ImportProfile

| Field | Required | Type |
|-------|:--------:|------|
| `firstName` | ✓ | string |
| `lastName` | ✓ | string |
| `phone` | ✓ | string |
| `customerType` | ✓ | string (enum) |
| `purchasedCategories` | ✓ | string[] |
| `job` | | string |
| `email` | | string |
| `address` | | string |
| `birthday` | | string (`YYYY-MM-DD`) |
| `marriageDate` | | string |
| `importantDate` | | string |
| `firstVisitDate` | | string |
| `customerAgeRange` | | string |
| `description` | | string |
| `signature` | | string |

---

## 6. Segments & status

### Segments (computed server-side)

| Segment | Rule |
|---------|------|
| `vip` | `isVip` is true |
| `new` | Zero orders |
| `returning` | Two or more orders |
| `inactive` | No activity in 90+ days (falls back to registration date) |
| `active` | Default for customers with orders who don't match above |

### Status mapping

| API `status` | Database `users.status` |
|--------------|-------------------------|
| `active` | `active` |
| `blocked` | `banned` |

### Blocked customer effects

When connected to storefront auth (planned):

- Cannot log in
- Cannot checkout
- Wishlist and cart disabled

---

## 7. Error codes

| HTTP | `error.code` | When |
|------|--------------|------|
| 401 | `UNAUTHORIZED` | Missing or invalid JWT |
| 403 | `FORBIDDEN` | Valid JWT but insufficient role |
| 404 | `NOT_FOUND` | Customer UUID not found |
| 409 | `PHONE_EXISTS` | Duplicate phone on create/update |
| 409 | `CONFLICT` | Duplicate email |
| 422 | `VALIDATION_ERROR` | Invalid input, block reason, tag, or UUID |
| 422 | `INVALID_JSON` | Malformed request body |
| 500 | `INTERNAL_ERROR` | Unexpected server failure |

---

## 8. Examples

### Quick search (name / phone)

```bash
curl -s "http://localhost:8080/api/v1/admin/customers?q=912&page=1&perPage=10" \
  -H "Authorization: Bearer $TOKEN"
```

### Advanced filters (CRM fields — `q` ignored)

```bash
curl -s "http://localhost:8080/api/v1/admin/customers?gender=female&customerTypes=vip&page=1&perPage=10" \
  -H "Authorization: Bearer $TOKEN"
```

### Advanced: find by customer ID or email

```bash
curl -s "http://localhost:8080/api/v1/admin/customers?id=550e8400-e29b-41d4-a716-446655440000" \
  -H "Authorization: Bearer $TOKEN"

curl -s "http://localhost:8080/api/v1/admin/customers?email=sara@example.com" \
  -H "Authorization: Bearer $TOKEN"
```

### Create quick-add customer

```bash
curl -s -X POST http://localhost:8080/api/v1/admin/customers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "importMode": "quick",
    "fullName": "Sara Mohammadi",
    "phone": "+989121234567",
    "locale": "fa"
  }'
```

### Block customer

```bash
curl -s -X POST "http://localhost:8080/api/v1/admin/customers/$CUSTOMER_ID/block" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "fraud_suspicion", "note": "Chargeback pattern"}'
```

### Get profile

```bash
curl -s "http://localhost:8080/api/v1/admin/customers/$CUSTOMER_ID" \
  -H "Authorization: Bearer $TOKEN"
```

---

## 9. Admin UI integration

The Next.js admin app calls these endpoints via `lib/api/customers/*`.

| Environment | `NEXT_PUBLIC_API_BASE_URL` | Resolved path |
|-------------|---------------------------|---------------|
| Local API | `http://localhost:8080/api/v1` | `{base}/admin/customers` |
| Production | `https://api.rehil.gallery/api/v1` | `{base}/admin/customers` |

No UI code changes are required when switching from the example in-memory API — only the env variable.

---

## 10. Saved views & segments

### GET `/segments`

Returns built-in segment counts (from `customer_profiles.segment`) and the current user's saved **segment** shortcuts.

```json
{
  "builtin": [
    { "segment": "vip", "label": "VIP", "count": 5001 },
    { "segment": "new", "label": "New", "count": 60000 }
  ],
  "saved": [
    {
      "id": "uuid",
      "name": "VIP high spenders",
      "viewType": "segment",
      "filters": { "customerTypes": ["vip"] },
      "isShared": false,
      "position": 0,
      "ownerId": "uuid",
      "createdAt": "2026-06-02T10:00:00Z",
      "updatedAt": "2026-06-02T10:00:00Z"
    }
  ]
}
```

### Saved views CRUD

`filters` mirrors list query params (`q`, `gender`, `customerTypes`, dates as `YYYY-MM-DD`, etc.). Apply a saved view by copying `filters` onto `GET /admin/customers`.

| `viewType` | Use case | Required in `filters` |
|------------|----------|------------------------|
| `segment` | Sidebar shortcut | `segment` (optional extra filters allowed) |
| `filter` | Advanced filter preset | At least one of `q` or advanced params |

Shared views (`isShared: true`) are visible to all staff; only the owner can update/delete.

```bash
# Save advanced filter
curl -s -X POST "http://localhost:8080/api/v1/admin/customers/saved-views" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "VIP customers",
    "viewType": "filter",
    "filters": { "customerTypes": ["vip"] },
    "isShared": true,
    "position": 1
  }'
```

---

## 11. Not yet implemented

| Feature | Notes |
|---------|-------|
| `GET /admin/customers/export` | CSV export with rate limiting |
| Fine-grained roles | `support`, `crm`, `analyst` permission matrix |
| OTP login + set-password | Customers with null `password_hash` |
| Export audit logging | Server-side export event |

---

## 12. Related documents

| Document | Description |
|----------|-------------|
| [migrations/000002_customer_crm.up.sql](../migrations/000002_customer_crm.up.sql) | CRM schema migration |
