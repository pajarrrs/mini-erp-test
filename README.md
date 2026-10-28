# Mini ERP Test - Inventory & Sales

A small app to manage products, adjust stock and record sales orders.
PostgreSQL → Go API → React + TypeScript UI.

```
mini-erp-test/
├── backend/
│   ├── cmd/api/main.go          start the HTTP server
│   ├── db/                      schema.sql, seed.sql
│   └── internal/
│       ├── model/               data structs (JSON shapes)
│       ├── apperror/            the one API error type
│       ├── validation/          request checks (no database)
│       ├── repository/          all SQL + transactions
│       ├── handler/             HTTP handlers and routes
│       └── testutil/            clean database for tests
├── frontend/
│   └── src/
│       ├── api/                 types.ts (shared types) + client.ts (all fetch calls)
│       ├── utils/               formatRupiah, timeAgo, form validation
│       ├── hooks/               useDebounce
│       └── components/          pages, modal, forms
├── docker-compose.yml
└── README.md
```

## How to run

### Everything with Docker (one command)

```bash
docker compose up --build
```

- App: http://localhost:5173
- API: http://localhost:8080/api/products

On the first start PostgreSQL runs `backend/db/schema.sql` (tables) and
`backend/db/seed.sql` (14 products, 2 orders). To reset the data:

```bash
docker compose down -v
```

### Without Docker for the apps (development)

```bash
docker compose up -d db                       # 1. PostgreSQL on port 5432
cd backend && go run ./cmd/api                # 2. API on port 8080
cd frontend && npm install && npm run dev     # 3. UI on port 5173
```

**Dev setup / CORS:** Vite forwards every `/api/...` request to the backend
(`frontend/vite.config.ts`). The browser only talks to `localhost:5173`, so the
backend needs no CORS middleware.

## Tests

```bash
docker compose up -d db
cd backend && go test -race ./...     # needs the database
cd frontend && npm test               # Vitest, no backend needed
```

Backend tests (real PostgreSQL). Every test creates its own temporary schema
(`test_...`) inside the `mini-erp-test` database and drops it afterwards, so the
app data is never touched:

| Test | What it proves |
|---|---|
| `TestCreateOrder_InsufficientStock_WritesNothing` | 409, and no order, item, movement or stock change is saved |
| `TestCreateOrder_HappyPath` | stock goes down, "out" movements are written, prices/total come from the DB |
| `TestCreateOrder_UnknownSKU` | 400 with the field name |
| `TestAdjustStock` | "in" writes a movement; overdraw returns 409 with `available`; 404 |
| `TestCancelOrder_RestoresStock` | stock comes back; second cancel is 409 |
| `TestDeleteProduct` | 409 `PRODUCT_HAS_HISTORY`, 204, 404 |
| `TestListProducts_SearchAndFilter` | search and low-stock filter |
| `TestConcurrentOrders_StockNeverNegative` | 30 orders at once on stock 10 → exactly 10 succeed, stock ends at 0 |
| `TestConcurrentAdjustments_StockNeverNegative` | 20 "out" adjustments at once on stock 5 → stock ends at 0 |
| `TestOrderFlowOverHTTP` | status codes and JSON error envelope |

Frontend tests: `formatRupiah`, `timeAgo`, quantity and order form validation,
error envelope reading, products URL building.

## API

Every error has the same shape:

```json
{ "error": { "code": "INSUFFICIENT_STOCK", "message": "Not enough stock for 006", "sku": "006", "available": 8 } }
```

| Method | Path | Result |
|---|---|---|
| GET | `/api/products?search=&status=&page=&limit=` | `{data, page, limit, total}` - status: `active`, `inactive`, `low_stock` |
| POST | `/api/products` | 201 · 400 · 409 `SKU_ALREADY_EXISTS` |
| PUT | `/api/products/{id}` | 200 · 400 · 404 |
| DELETE | `/api/products/{id}` | 204 · 404 · 409 `PRODUCT_HAS_HISTORY` |
| POST | `/api/products/{id}/stock` | 200 · 400 · 404 · 409 `INSUFFICIENT_STOCK` |
| GET | `/api/products/{id}/movements` | `{data: [...]}`, newest first |
| GET | `/api/summary` | numbers for the summary cards |
| GET | `/api/orders` | latest 50 orders |
| POST | `/api/orders` | 201 · 400 · 409 `INSUFFICIENT_STOCK` |
| GET | `/api/orders/{id}` | order + items |
| POST | `/api/orders/{id}/cancel` | 200 · 404 · 409 `ORDER_ALREADY_CANCELLED` |

### curl examples

```bash
curl "localhost:8080/api/products?search=beras&page=1&limit=10"

curl -X POST localhost:8080/api/products -H 'Content-Type: application/json' \
  -d '{"sku":"015","name":"Minyak Kelapa","price":32000}'

curl -X PUT localhost:8080/api/products/1 -H 'Content-Type: application/json' \
  -d '{"sku":"001","name":"Beras","price":79000}'

curl -X POST localhost:8080/api/products/1/stock -H 'Content-Type: application/json' \
  -d '{"type":"in","qty":10,"note":"Delivery PO-3001"}'

curl -X POST localhost:8080/api/products/1/stock -H 'Content-Type: application/json' \
  -d '{"type":"out","qty":99999}'
# -> 409 {"error":{"code":"INSUFFICIENT_STOCK","available":...}}

curl localhost:8080/api/products/1/movements

curl -X POST localhost:8080/api/orders -H 'Content-Type: application/json' \
  -d '{"customer":"Toko Maju Jaya","items":[{"sku":"001","qty":2},{"sku":"003","qty":5}]}'

curl localhost:8080/api/orders/1
curl -X POST localhost:8080/api/orders/1/cancel

curl -X DELETE localhost:8080/api/products/1
# -> 409 PRODUCT_HAS_HISTORY
```

## Design decisions

**Stock can never go negative, even with many requests at the same time.**
- Creating an order is one transaction. First it locks all products of the order
  with `SELECT ... FOR UPDATE` (always in `id` order, so two orders cannot deadlock).
  Then it checks every item. Only when everything is OK does it write the order,
  the items, the movements and lower the stock. Any error → `ROLLBACK`, nothing changes.
- Every stock decrease also has the check inside the UPDATE:
  `UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1`.
  If 0 rows changed → 409. We never do "SELECT stock, then UPDATE" without a lock.
- The database has `CHECK (stock >= 0)` as a last safety net.
- Cancel uses `UPDATE ... WHERE status = 'CREATED'`, so a double click can only
  give the stock back once.

**`last_movement_at` is calculated, not stored.** It is `MAX(created_at)` from
`stock_movements` for that product (with an index). It can never be out of date.

**Stock only changes through movements.** `PUT /api/products/{id}` cannot change
stock. Adjustments, orders and cancels always write a movement row in the same
transaction, so stock always equals the sum of the movements.

**History is never deleted.** A product with movements cannot be deleted (409).
Its SKU also cannot be renamed if it is used in an order (the foreign key blocks it → 409).

**Prices come from the database.** The order request only has `sku` and `qty`.
`unit_price` and `total` are calculated by the server. Money is whole Rupiah (integer).

**Frontend.** No router or state library: `App.tsx` keeps the current page in state.
All HTTP calls are in `src/api/client.ts`, which also turns error responses into
an `ApiError`. After a change (stock adjustment, new order, cancel) `App` increases
`refreshKey`, and every component that shows stock loads fresh data from the server.

## What I would improve with more time

- Screens to create, edit and delete products (the API already supports it).
- Movement history per product in the UI (the endpoint exists).
- `Idempotency-Key` on `POST /api/orders` so a retried request never creates two orders.
- Component tests (React Testing Library) for the modal and the order form.
- A production build for the frontend (static files behind nginx) instead of the dev server.

## Hours spent

- 6 Hours
