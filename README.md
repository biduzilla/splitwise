# Rateio

A multi-user expense splitter built as a Go microservices backend with an Angular SPA frontend. Groups record shared expenses, the system computes individual balances, and suggests the minimum number of transfers to settle all debts.

---

## Architecture

```
                     ┌──────────────────────────────────────────┐
                     │            Angular SPA (web/)            │
                     │   Port 4200 · cookies httpOnly · proxy   │
                     └───────────────┬──────────────────────────┘
                                     │  /api/{auth,groups,expenses,settlements,balances}
                                     ▼
       ┌──────────────┬──────────────┬──────────────┬──────────────┐
       │   ms_auth    │   ms_group   │  ms_expense  │  ms_balance  │
       │   :4001      │   :4002      │   :4003      │   :4004      │
       │  Users/JWT   │ Groups/Members│ Expenses    │ Balances/    │
       │  Refresh     │ Invitations  │  Shares      │ Settlements  │
       └──────┬───────┴──────┬───────┴──────┬───────┴──────┬───────┘
              │              │              │              │
              │     REST (synchronous, between services)    │
              │              │              │              │
              ▼              ▼              ▼              ▼
         ┌─────────────────────────────────────────────────────┐
         │  PostgreSQL  ·  Redis (cache)  ·  Kafka (events)    │
         │  Jaeger (traces) · Prometheus · Grafana             │
         └─────────────────────────────────────────────────────┘
```

**Communication patterns:**
- **REST synchronous** for direct queries between services (e.g. `ms_expense` → `ms_group` to validate membership)
- **Kafka asynchronous** for event-driven recalculation (e.g. `ms_expense` publishes `expense.created` → `ms_balance` consumer recalculates balances)

---

## Tech Stack

### Backend
- **Go 1.25** · Chi router · `encoding/json/v2`
- **PostgreSQL** with one schema per service (`auth`, `grp`, `exp`, `bal`)
- **Redis** for response caching and idempotency locks
- **Kafka** (Sarama) for event-driven integration
- **OpenTelemetry** traces exported to Jaeger via OTLP
- **Prometheus** metrics + Grafana dashboards

### Frontend
- **Angular 22** (signal-first, standalone, zoneless)
- **Tailwind CSS 4** utility-first design system
- **Signal Forms** (`@angular/forms/signals`) — type-safe reactive forms
- **HttpClient** with functional interceptors (auth refresh, error mapping)
- `provideAppInitializer` bootstraps session via `GET /v1/auth/session`

---

## Repository Layout

```
splitwise/
├── backend/
│   ├── go.work                         # workspace with 6 modules
│   ├── docker-compose.yml              # postgres, redis, kafka, jaeger, prometheus, grafana
│   ├── shared/                         # shared Go module
│   │   ├── apierror/                   # uniform error envelope + ErrorHandler
│   │   ├── auth/                       # JWT RS256, httpOnly cookies, UserContext
│   │   ├── cache/                      # Redis adapter + KeyBuilder + FetchOrCache
│   │   ├── db/                         # transaction manager + SQL helpers
│   │   ├── filters/                    # pagination, sort, Page[T], Map
│   │   ├── httpx/                      # httpjson (ReadJSON) + httputils (Respond, ParseUUID, GetFilters)
│   │   ├── models/                     # BaseModel (audit columns)
│   │   ├── obs/                        # logger, otel, metrics, middleware stack
│   │   ├── transaction/               # RunInTx + WriteExecutor (cache invalidation)
│   │   └── validation/                 # custom validator
│   ├── messaging/                      # Kafka producer/consumer with OTel propagation
│   ├── ms_auth/                        # users, refresh tokens, JWT login/logout
│   ├── ms_group/                       # groups, memberships, invitations (JWT-signed links)
│   ├── ms_expense/                     # expenses, shares, split equal/exact
│   └── ms_balance/                     # balances, settlements, minimize algorithm, Kafka consumer
└── web/
    ├── proxy.conf.json                 # /api/{auth,groups,expenses,settlements,balances} → services
    └── src/app/
        ├── core/                       # services, guards, interceptors, models, context
        ├── features/                   # auth, groups, expenses, balances
        └── shared/                     # layout, components, ui (design system)
```

---

## Backend Services

| Service  | Port | Schema | Responsibility |
|----------|------|--------|----------------|
| `ms_auth` | 4001 | `auth` | Users, signup, login, JWT (access 15min / refresh 30d with family rotation), session check |
| `ms_group` | 4002 | `grp` | Groups, memberships (max 10 per group), invitations (JWT-signed, 72h, single-use) |
| `ms_expense` | 4003 | `exp` | Expenses, expense_shares, split modes (equal/exact), idempotent POST, soft delete |
| `ms_balance` | 4004 | `bal` | Materialized group balances, settlement suggestions (greedy minimize), Kafka consumer recalculates on events |

### Cross-cutting concerns

- **Soft delete** (`deleted = true` column) on every mutable entity
- **Optimistic locking** via `version` column — non-matching `PUT` returns 409
- **Idempotency** — `POST /expenses` and `POST /settlements` accept `Idempotency-Key` header, deduped via `*.idempotency_keys` table + Redis lock
- **Multi-tenancy** — `RequireGroupMember` middleware injects `groupID` into context; no query runs without `WHERE group_id = ?`
- **Money in cents** — `BIGINT`, never `float` (R$ 33,33 = 3333)
- **Audit columns** — `created_by`, `updated_by`, `created_at`, `updated_at` on every entity

### Event Catalog (Kafka topic `rateio.events`)

| Event | Producer | Consumer | Purpose |
|---|---|---|---|
| `expense.created` | ms_expense | ms_balance | Trigger balance recalculation |
| `expense.updated` | ms_expense | ms_balance | Recalculate |
| `expense.deleted` | ms_expense | ms_balance | Recalculate |
| `settlement.created` | ms_balance | ms_balance (self) | Recalculate |
| `settlement.deleted` | ms_balance | ms_balance (self) | Recalculate |
| `group.*` | ms_group | (future, audit) | Audit trail |

---

## Frontend Structure

```
src/app/
├── core/
│   ├── context/group-context.service.ts   # active group as signal
│   ├── guard/{auth,group-member}.guard.ts
│   ├── interceptors/{auth,error}.interceptor.ts
│   ├── models/api.types.ts                # DTOs matching Go 1:1
│   └── services/{auth,group,expense,balance}.service.ts
├── features/
│   ├── auth/{login,signup,accept-invite}/
│   ├── groups/{groups-list,group-form,group-detail}/
│   ├── expenses/{expenses-list,expense-form}/
│   └── balances/{balances-view,settlement-form}/
└── shared/
    ├── components/{pagination,avatar,empty-state,loading-spinner}/
    ├── layout/shell.component.ts          # nav + active group + logout
    └── ui/styles.ts                       # Tailwind design system constants
```

### Key UX patterns

- **URL as source of truth** — filters, pagination, sort live in query params (F5-safe, shareable)
- **Active group in URL** — `/groups/:groupId/*` enables multi-tab workflows
- **Session bootstrap** — `provideAppInitializer` calls `GET /v1/auth/session` before router activates, preventing flash-to-login on F5
- **Refresh token serialization** — concurrent 401s share a single refresh request via `shareReplay`
- **Idempotency** — every critical POST sends `Idempotency-Key: crypto.randomUUID()` to prevent duplicate creates on double-click

---

## API Endpoints

### Auth (`ms_auth`)
| Method | Path | Description |
|---|---|---|
| POST | `/v1/users` | Signup |
| POST | `/v1/auth` | Login (sets httpOnly cookies) |
| POST | `/v1/auth/refresh` | Rotate tokens |
| POST | `/v1/auth/logout` | Revoke family + clear cookies |
| GET | `/v1/auth/session` | `{ authenticated, user_id?, email? }` (always 200) |

### Group (`ms_group`)
| Method | Path | Description |
|---|---|---|
| POST | `/v1/groups` | Create group |
| GET | `/v1/groups` | List user's groups (paginated) |
| GET | `/v1/groups/{id}` | Group detail with members |
| DELETE | `/v1/groups/{id}` | Delete (owner only) |
| POST | `/v1/groups/{id}/invitations` | Generate invite link (JWT, 72h) |
| POST | `/v1/invitations/{token}/accept` | Accept invite |
| DELETE | `/v1/groups/{id}/members/{userId}` | Remove member (owner only) |

### Expense (`ms_expense`)
| Method | Path | Description |
|---|---|---|
| POST | `/v1/expenses` | Create (Idempotency-Key) |
| GET | `/v1/expenses?group_id=&page=&sort=&start_date=&end_date=` | List |
| GET | `/v1/expenses/{id}` | Detail |
| PUT | `/v1/expenses/{id}` | Update (version required) |
| DELETE | `/v1/expenses/{id}` | Soft delete (creator or owner) |

### Balance & Settlement (`ms_balance`)
| Method | Path | Description |
|---|---|---|
| GET | `/v1/settlements/groups/{id}/balances` | Group balances |
| GET | `/v1/settlements/groups/{id}/suggested` | Minimized transfer suggestions |
| POST | `/v1/settlements` | Register settlement (Idempotency-Key) |
| GET | `/v1/settlements?group_id=` | List settlements |
| DELETE | `/v1/settlements/{id}` | Soft delete |

---

## Quick Start

### Prerequisites
- Go 1.25+
- Node 22 LTS
- Docker + Docker Compose

### Backend

```bash
cd backend
cp .env.example .env          # adjust passwords/ports if needed
docker compose up -d          # postgres, redis, kafka, jaeger, prometheus, grafana

# Run services (4 terminals, or use make targets)
go run ./ms_auth/cmd/api
go run ./ms_group/cmd/api
go run ./ms_expense/cmd/api
go run ./ms_balance/cmd/api
```

### Frontend

```bash
cd web
npm install
ng serve                      # http://localhost:4200
```

### Smoke test

```bash
# Signup
curl -i -X POST http://localhost:4001/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@example.com","password":"Senha123A","name":"Ana"}'

# Login (saves cookies)
curl -i -c cookies.txt -X POST http://localhost:4001/v1/auth \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@example.com","password":"Senha123A"}'

# Session check
curl -b cookies.txt http://localhost:4001/v1/auth/session
# → {"authenticated":true,"user_id":"...","email":"ana@example.com"}
```

---

## Key Design Decisions

### Cookies httpOnly over localStorage
Tokens never touch JavaScript. `ms_auth` sets `Set-Cookie: access_token=...; HttpOnly; SameSite=Lax`. The browser attaches cookies automatically to every API call. The SPA can't read the token — it asks `GET /v1/auth/session` on boot to know who's logged in.

### Shared database, separate schemas
All 4 services share one Postgres instance but each has its own schema (`auth`, `grp`, `exp`, `bal`). This allows JOINs across services for now (e.g. `ms_group` enriches memberships with user names) while keeping tables isolated. The prefix convention (`auth_users`, `grp_groups`) prepares for splitting into separate databases later.

### Materialized balances + event-driven recalculation
`bal.group_balances` is a materialized table, not computed on every read. When `ms_expense` creates/updates/deletes an expense, it publishes a Kafka event. The `ms_balance` consumer consumes it, fetches all expenses + settlements for that group, recomputes balances, and upserts the table. Reads are O(1) cache lookups.

### Greedy minimize algorithm for settlement suggestions
Given a balance map, the algorithm splits members into creditors (balance > 0) and debtors (balance < 0), sorts by magnitude descending, then greedily matches the largest debtor with the largest creditor. Complexity: O(n log n). Result: minimum number of transfers to zero all balances.

### Signal-first Angular
No `BehaviorSubject`, no `ReactiveFormsModule`. State is `signal()` everywhere. Forms use `@angular/forms/signals` with `form()`, `[formField]`, `required()`, `min()`. Templates call signals with `()` — no pipe needed.

---

## Observability

- **Jaeger** at http://localhost:16686 — distributed traces from HTTP request through Kafka consumer
- **Prometheus** at http://localhost:9090 — RED metrics (Rate, Errors, Duration)
- **Grafana** at http://localhost:3001 — dashboards (empty by default, provisioning in `grafana/`)
- **Structured JSON logs** with `request_id`, `service.name`, `trace_id`

OTel context propagates through Kafka message headers (`traceparent`), so a single trace spans `ms_expense` producer → Kafka → `ms_balance` consumer.

---

## Roadmap

Done:
- [x] Auth with refresh rotation
- [x] Groups, memberships, invitations
- [x] Expenses with split equal/exact
- [x] Balances + settlement suggestions
- [x] Settlement registration with idempotency
- [x] Frontend shell with all features

In progress / next:
- [ ] TTL on Redis cache (currently keys are eternal — recommend 5min default)
- [ ] Membership validation on settlement creation (`from`/`to` must be group members)
- [ ] `RequireGroupMember` middleware on `DELETE /settlements/{id}`
- [ ] Pagination on `repo.FindByGroupID` (currently in-memory slice)
- [ ] End-to-end tests (Playwright)
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Production deploy (Dockerfiles + nginx reverse proxy + HTTPS)

---

## License

Private project. Contact the maintainer for access.
