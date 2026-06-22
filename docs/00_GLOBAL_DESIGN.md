# AhorraApp — Global Design Document (Single Source of Truth)

> Version 1.0 · Minimum Viable Product (MVP) · Initial market: Venezuela
> This document summarizes ALL agreed architecture, stack, and infrastructure decisions.
> It is the root input for Spec-Driven Development. Read it before any other file.

---

## 1. Product Vision

**Problem:** In high-inflation contexts (Venezuela), supermarket prices vary widely between
stores and over time. Shoppers have no way to know where to buy cheaper.

**Solution:** A mobile app where users photograph their receipt. The app extracts (OCR) the
products, prices, and **the store/merchant**, records them, and with the aggregated volume of
many users generates **average prices per product and per store**. This lets the app recommend
where to do the next shopping trip more cheaply.

**Value loop (flywheel):**
Users upload receipts → price data → useful recommendations → more users → more data →
better recommendations. Gamification (points per receipt) accelerates the loop.

**Future monetization (NOT in the MVP, only the architecture is prepared):**
- API for supermarkets: publish real inventory and offers.
- Subscription plans (defined in a later stage).

---

## 2. Agreed Technical Decisions (and why)

| Area | Decision | Rationale |
|------|----------|-----------|
| **Mobile app** | **Flutter (Dart)** | Single iOS/Android codebase, high performance (compiles to native), excellent for camera UI. |
| **Backend** | **Go (Golang)** + **Clean Architecture** | Most efficient language in CPU/RAM per dollar; single binary, instant startup, ideal for a small VPS and for orchestrating a concurrent OCR pipeline. |
| **Database** | **PostgreSQL 16** | Robust relational DB, supports JSONB for semi-structured receipt data and geospatial extensions (PostGIS) to locate stores. |
| **Cache / queue** | **Redis** | OCR job queue (async processing) and average-price cache. |
| **OCR (Phase 1, free)** | **PaddleOCR self-hosted** (Python microservice) behind an abstract `OCRProvider` interface | Zero per-scan cost while developing and validating. Strong multilingual performance. |
| **OCR (Phase 2+, paid)** | Mindee / Google Document AI / Tabscanner | Migrate by swapping ONE implementation of the interface, without touching the rest of the system. |
| **Image storage** | **MinIO (S3-compatible) self-hosted in Docker** for development; Hetzner Object Storage for production | Same S3 API in both, so the `StorageProvider` adapter is unchanged — only the endpoint URL differs. Zero cost and zero accounts locally. |
| **Infrastructure** | **Hetzner Cloud + Coolify** (self-hosted PaaS) | Best price/performance in 2026. Coolify gives a Heroku-like experience (git deploy, automatic SSL) on a cheap VPS. |
| **Methodology** | **GitHub Spec Kit** (Spec-Driven Development) | 2026 standard, agent-agnostic; officially supports Windsurf and OpenCode. |
| **Editor / AI agent** | **OpenCode** (with OpenCode Go subscription) | Uses a subscription you already pay for; lets you switch models via `/models`. Windsurf stays compatible as an optional fallback. |

### Key design principle: **OCR is a replaceable detail**
Thanks to Clean Architecture, OCR lives behind a *port* (interface). We start with the free
option and migrate to a paid one when volume and revenue justify it, without rewriting the
business logic.

### Key design principle: **Local-first development, cloud only when publishing**
The entire stack runs on your own machine with Docker — no external accounts, no cloud bills —
during development. Every external dependency has a local equivalent that speaks the same
protocol, so moving to the cloud later is a configuration change, not a rewrite:

| Concern | Local (development) | Cloud (production, later) | What changes when you move |
|---------|--------------------|--------------------------|----------------------------|
| Database | PostgreSQL in Docker | PostgreSQL on the VPS (or managed) | Connection string only |
| Cache/queue | Redis in Docker | Redis on the VPS | Connection string only |
| Image storage | **MinIO in Docker** (S3 API) | Hetzner Object Storage (S3 API) | Endpoint URL + keys only |
| OCR | PaddleOCR in Docker | Same container on the VPS | Nothing (same image) |
| Hosting | `docker compose up` on your laptop | Hetzner + Coolify | Deploy step, not code |
| Source control | Local Git | GitHub (optional backup) | `git remote add` once |

This means **Phase 3 (cloud deployment) is optional and deferred**. You build and demo the full
MVP locally for $0, and only provision the cloud when you want other people to use the app.

---

## 3. System Architecture (MVP)

```
┌─────────────────────┐
│   Flutter App       │  (camera, receipt capture, EDITABLE summary
│   (iOS / Android)   │   screen, gamification, ranking)
└──────────┬──────────┘
           │ HTTPS / REST (JSON)
           ▼
┌─────────────────────────────────────────────┐
│   Go Backend API (Clean Architecture)        │
│  ┌────────────┬──────────────┬─────────────┐ │
│  │ Handlers   │ Use cases    │ Repositories│ │
│  │ (HTTP)     │ (domain)     │ (Postgres)  │ │
│  └────────────┴──────────────┴─────────────┘ │
│         │                          │          │
│         ▼ enqueue                  ▼          │
│   ┌──────────┐              ┌────────────┐   │
│   │  Redis   │◄─worker──────│ PostgreSQL │   │
│   │ (queue)  │              │ + PostGIS  │   │
│   └────┬─────┘              └────────────┘   │
└────────┼──────────────────────────────────────┘
         │ calls (internal HTTP)
         ▼
┌─────────────────────┐         ┌──────────────────────────┐
│ OCR Service         │         │ Object Storage (S3 API)  │
│ (PaddleOCR, Python) │◄────────│ MinIO locally /          │
│ behind OCRProvider  │         │ Hetzner Obj. Storage prod│
└─────────────────────┘         └──────────────────────────┘
```

> All boxes above run as containers in a single `docker compose up` on your machine during
> development. Nothing here requires a cloud account.

### "Upload receipt" flow
1. User takes a photo → app uploads it to the backend.
2. Backend stores the image in Object Storage and creates a `receipt` record in `PENDING` state.
3. Backend enqueues a job in Redis.
4. A Go *worker* picks up the job, calls the OCR service, receives structured text.
5. The backend parses: **store/merchant**, date, list of {product, quantity, price}.
6. The receipt moves to `NEEDS_REVIEW` and an **editable text summary** is returned to the app.
7. The user **corrects and confirms**. On confirmation:
   - Prices are normalized and fed into the averaging engine.
   - Loyalty points are awarded.
   - The receipt moves to `CONFIRMED`.
8. The price engine recomputes averages per (product × store) and feeds the
   "where to buy cheaper" ranking.

---

## 4. Data Model (MVP core entities)

- **User**: id, phone/email, displayName, points, createdAt.
- **Store (merchant)**: id, name, branch, address, lat/long (PostGIS), createdAt.
- **Receipt**: id, userId, storeId, imageUrl, status (PENDING/NEEDS_REVIEW/CONFIRMED/REJECTED), purchaseDate, total, createdAt.
- **ReceiptItem (line item)**: id, receiptId, rawText, productId, quantity, unitPrice, lineTotal.
- **Product (canonical product)**: id, normalizedName, category, unit (kg, lt, unit).
- **PriceObservation**: id, productId, storeId, unitPrice, currency, observedAt, receiptId.
- **PriceAggregate (average)**: productId, storeId, avgPrice, minPrice, sampleCount, lastUpdated.
- **LoyaltyTransaction**: id, userId, points, reason, createdAt.

> Currency note: in Venezuela both Bs. and USD coexist. The `currency` field is mandatory on
> every price observation. Aggregations are computed per currency.

---

## 5. Infrastructure and Costs

### Phase A — Local development (current stage): **$0/month, no accounts**

Everything runs on your machine through one `docker compose up`:

| Component | How it runs locally | Cost |
|-----------|--------------------|------|
| Go API | Docker container | $0 |
| PostgreSQL + PostGIS | Docker container | $0 |
| Redis | Docker container | $0 |
| MinIO (S3-compatible image storage) | Docker container | $0 |
| PaddleOCR service | Docker container | $0 |
| Source control | Local Git | $0 |
| **Total during development** | | **$0** |

This is the recommended stage right now. You can build, run, and demo the entire MVP locally
without creating a single cloud account. GitHub is optional (only as an off-machine backup).

### Phase B — Cloud deployment (later, optional): target $50–$100/month

Only when you want other people to use the app over the internet:

| Resource | Provider | Spec | Approx. cost |
|----------|----------|------|--------------|
| Main VPS (Go API + Coolify + Postgres + Redis) | Hetzner CX32 | 4 vCPU / 8 GB RAM / 80 GB | ~$8/mo |
| OCR VPS (PaddleOCR, separate to isolate CPU) | Hetzner CX22 | 2 vCPU / 4 GB RAM | ~$4/mo |
| Object Storage (images) | Hetzner | 1 TB | ~$6/mo |
| Automatic backups | Hetzner | snapshots | ~$2/mo |
| Domain + email | Namecheap/Cloudflare | — | ~$2/mo |
| **Infrastructure subtotal** | | | **~$22/mo** |
| Headroom for spikes / paid OCR after validation | | | rest of budget |

> Production starts WELL below $50. The margin ($50–$100) is a buffer to: scale the VPS,
> enable paid OCR, or add a second node when load grows.
> Scaling strategy: vertical first (bump the VPS plan), then horizontal (split the OCR worker,
> Postgres read replica, load balancer).

**Moving from Phase A to Phase B changes configuration, not code:** the same Docker images and
the same `StorageProvider`/database interfaces point to cloud endpoints via environment
variables. Nothing in the domain or use-case layers changes.

**"Zero-ops" alternative** if you prefer not to administer servers: Railway or Render
(~$5–20/mo per service). More expensive at scale but less operational work. Recommendation:
Hetzner + Coolify for the budget and the learning; migration is simple because everything runs
in Docker.

---

## 6. Reference Success Cases (researched)

- **Receipt Hog / Receipt Hog-style apps:** rewards model for uploading receipts; validates that
  gamification drives data volume. Lesson: the incentive must be immediate and visible.
- **Mindee / Tabscanner / Google Document AI:** confirm that production receipt OCR uses
  transformer models; the sensible path is *start self-hosted and migrate to an API* once volume
  passes the point where self-hosting engineering cost exceeds API cost (2026 references place
  this below ~100k scans/month).
- **T3 / 2026 startup stacks:** confirm PostgreSQL as the default database and the trend toward
  simple infra (Hetzner/Railway/Render) in the early stage.
- **GitHub Spec Kit (90k+ stars, 30+ agents):** de facto standard for building with AI in a
  disciplined way; supports Windsurf and OpenCode.

---

## 7. What YOU Do by Hand vs. What You Delegate to the AI

### By hand (human) — decisions and setup
1. Install local tools: Flutter SDK, Go, Docker, uv (Python), Specify CLI, OpenCode (Windsurf optional).
2. Run `specify init` and choose your agent.
3. Review and **approve** each spec/plan before the AI implements (this is the heart of SDD).
4. Test the app on your phone and validate OCR quality with real Venezuelan receipts.
5. (Local stage: no accounts needed. GitHub optional as backup.)
6. **Only when publishing:** create the Hetzner account + domain, provision the VPS, point DNS (one time).

### Delegated to the AI (via Spec Kit prompts) — implementation
- Generate the project constitution (`/constitution`).
- Write specs per feature (`/specify`).
- Generate technical plans (`/plan`).
- Break down into tasks (`/tasks`).
- Implement task by task (`/implement`).
- Write tests, Dockerfiles, SQL migrations, and the Flutter app.

---

## 8. MVP Build Order (epics)

1. **E1 — Go backend skeleton** (Clean Architecture, health-check, Docker, Postgres, migrations).
2. **E2 — User auth** (phone or email register/login, JWT).
3. **E3 — Receipt upload + image storage + PENDING state.**
4. **E4 — OCR service (PaddleOCR) + `OCRProvider` interface + async worker.**
5. **E5 — Receipt parsing: store, date, line items; NEEDS_REVIEW state.**
6. **E6 — Flutter app: camera, upload, EDITABLE summary screen, confirmation.**
7. **E7 — Price engine: observations → averages per product×store.**
8. **E8 — "Where to buy cheaper" ranking + product search.**
9. **E9 — Gamification: points per confirmed receipt + rewards screen.**
10. **E10 — Hardening: rate limiting, validation, observability, backups.**

> The "demonstrable" MVP is reached when E6 is done. E7–E9 turn it into the product with real value.

---

## 9. Tuning State (what's already agreed)

- ✅ Mobile: Flutter
- ✅ Backend: Go + Clean Architecture
- ✅ DB: PostgreSQL (+ PostGIS to geolocate stores)
- ✅ OCR: start free (PaddleOCR self-hosted) behind a replaceable interface
- ✅ Infra: **local-first ($0, no accounts) via Docker** during development; Hetzner + Coolify (~$22/mo) only when publishing
- ✅ Image storage: MinIO locally (S3 API), same adapter as Hetzner Object Storage in production
- ✅ Methodology: GitHub Spec Kit (SDD)
- ✅ Editor/Agent: **OpenCode (with OpenCode Go subscription)** as the chosen agent. The full
  step-by-step setup and build flow is in `01_GETTING_STARTED.md`. (Windsurf remains compatible
  as an optional fallback, since everything is file-based.)
- ✅ Reading the store/merchant from the receipt: included in E5
- ✅ Editable text summary before confirming: included in E5/E6
- ✅ Gamification: included in E9
- ✅ API for supermarkets and subscriptions: out of MVP, architecture prepared
- ✅ Working language/convention: **English** (code, specs, docs, commits)
- ✅ Development OS: **Windows native** (no WSL required — Spec Kit ships PowerShell scripts;
  Flutter's Android tooling works better on native Windows). macOS/Linux also fully supported.

### Pending decisions for the next stage
- Final authentication method (SMS OTP, which has a cost, vs. free email/password?).
- Target paid OCR provider for Phase 2.
- Definition of subscription plans and pricing.
- Product-name normalization strategy (key to good averages).
