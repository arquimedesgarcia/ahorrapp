<!--
============================================================================
SYNC IMPACT REPORT
============================================================================
Constitution version: (none) -> 1.0.0   [inaugural formal ratification]

Modified principles (title unchanged; wording strengthened):
  - Article I  - Clean Architecture in the Go backend
      Point 3 now explicitly names StorageProvider alongside OCRProvider as
      a required replaceable port (parity with 00_GLOBAL_DESIGN.md, which
      defines both as S3/OCR swap points). Point 4 restructured to list both
      named ports and their swap rules.
  - Article III - Tests
      Wording tightened to MUST; TDD intent clarified (tests written before
      the code they verify whenever the dependency direction allows).
  - Article V  - Data, currency, and normalization (was "Data and currency")
      Store now explicitly carries geolocation (lat/long via PostGIS).
      Currency rule unchanged. Normalization strategy explicitly deferred to
      spec/plan (was implicit).
  - Article VI - Simplicity, cost, and local-first (was "Simplicity and cost")
      New point 4: local-first development mandate, cloud deferred - drawn
      from 00_GLOBAL_DESIGN.md "Local-first development" key principle.

Added sections:
  - Governance (expanded): amendment procedure, semantic-versioning policy,
    compliance review expectations, mirror-maintenance note.
  - Version / Ratified / Last Amended footer line.

Removed sections: none.

Templates requiring updates:
  - .specify/templates/plan-template.md      - OK no change needed
      Its "Constitution Check" gate is generic
      ("[Gates determined based on constitution file]"); no stale article
      references to fix.
  - .specify/templates/spec-template.md       - OK no change needed
      No article-specific references; already uses MUST language.
  - .specify/templates/tasks-template.md      - OK no change needed
      No article-specific references.
  - .specify/templates/constitution-template.md - OK no change (template)
  - .opencode/commands/*                      - VERIFIED no change needed
      Directory now exists (11 command files). All references use
      `constitution.md` (lowercase) and contain no article-number
      references that would go stale. No outdated references found.

Runtime docs requiring updates:
  - docs/03_CONSTITUTION.md                  - UPDATED
      Mirror kept in sync with .specify/memory/constitution.md (both now
      carry v1.0.0).
  - docs/02_PROMPTS.md (PROMPT 1)            - RESOLVED
      Reworded "eight articles" -> "nine articles"; list now covers all
      nine articles including StorageProvider (Art. I), local-first (Art. VI),
      ready-to-grow (Art. VIII), and English (Art. IX).
  - docs/01_GETTING_STARTED.md (line 276)    - RESOLVED
      Inline list aligned with the nine articles: OCR folded into Art. I,
      Article VIII (ready to grow) added, Article IX (English) retained.

Deferred items / TODOs: none.
  RATIFICATION_DATE is set to 2026-06-21 as the first formal ratification.
  The original informal adoption date of the pre-existing unversioned
  constitution is unknown and not recoverable from the repo, so today's
  date is used as the baseline.

Flagged ambiguities (detailed in the final summary to the user):
  1. "Eight articles" (PROMPT 1) vs nine articles (actual file +
     getting-started line 276). Kept nine; flagged PROMPT 1 for rewording.
  2. OCR is listed as a separate principle in PROMPT 1 / getting-started but
     folded into Article I.4 in the constitution. Kept folded; both named
     ports (OCRProvider, StorageProvider) are now explicit in Article I.
  3. docs/03_CONSTITUTION.md and .specify/memory/constitution.md are
     intentionally duplicated. Both updated; recommend a single source of
     truth (or a generation/symlink step) to prevent future drift.
  4. File-name casing: VERIFIED — the file is already `constitution.md`
     (lowercase) on disk and in git's index. No rename needed.
============================================================================
-->

# AhorraApp Constitution

> Non-negotiable principles that EVERY spec, plan, task, and piece of code
> must respect. Spec Kit reads this file in every phase; if a plan violates
> a principle, it must be corrected.
>
> Source of truth: `docs/00_GLOBAL_DESIGN.md`. This constitution refines
> that document's decisions into binding articles. Where the global design
> and this constitution differ, this constitution prevails on governance;
> where the global design adds detail not covered here, it guides
> implementation without overriding these articles.

## Article I — Clean Architecture in the Go backend
1. The backend is organized in layers with dependencies pointing inward:
   `handlers (HTTP) -> usecases (domain/application) -> entities (pure domain)`,
   with `repositories` as interfaces defined in the domain and implemented
   in the outer (adapter) layer.
2. The domain (entities and usecases) MUST NOT import frameworks, database
   drivers, HTTP routers, or any infrastructure package.
3. Every external dependency (Postgres, Redis, OCR, Object Storage) MUST be
   accessed through an interface (port) defined in the domain; concrete
   implementations are adapters in the outer layer.
4. **Replaceable details — named ports.** At minimum, two ports MUST exist
   and be the only sanctioned way to reach their concern:
   - `OCRProvider` — OCR is a replaceable detail. Switching from the initial
     self-hosted PaddleOCR adapter to a paid API (Mindee / Google Document
     AI / Tabscanner) MUST NOT require changes outside its adapter.
   - `StorageProvider` — object storage is a replaceable detail. The same
     S3-compatible adapter MUST target MinIO in development and Hetzner
     Object Storage in production; only endpoint URL and credentials differ
     (via environment variables), with no change to the domain or use cases.

## Article II — Spec first, code second
1. No production code is written without an approved spec and an approved
   plan.
2. The spec describes WHAT and WHY (behavior, acceptance criteria), never
   the HOW.
3. The plan describes the HOW (technical decisions) and MUST cite the
   articles of this constitution that each decision satisfies.

## Article III — Tests
1. Every domain use case MUST have unit tests.
2. Critical endpoints (upload receipt, confirm receipt, price ranking) MUST
   have integration tests.
3. A TDD-style cycle is followed where practical: the spec defines the
   acceptance criteria, the tests prove them, and implementation follows.
   Tests are written before the code they verify whenever the dependency
   direction allows.

## Article IV — Explicit, versioned contracts
1. The API is REST/JSON with versioned contracts under `/api/v1/...`.
   Breaking changes require a new path version (`/api/v2/...`).
2. Every endpoint MUST document its request, response, and error shapes.
   Generating and maintaining an OpenAPI description is the preferred way.
3. The Flutter app consumes ONLY published contracts; it MUST NOT assume
   unspecified data shapes or undocumented fields.

## Article V — Data, currency, and normalization
1. Every `PriceObservation` MUST include a mandatory `currency` (Bs. or
   USD). Averages are computed per currency; currencies are NEVER mixed
   within a single average.
2. The store/merchant (`Store`) MUST be extracted and identified on every
   receipt. A Store records name, branch/address, and geolocation
   (lat/long, via PostGIS) when available.
3. Product names MUST be normalized to a canonical `Product` (with unit)
   before averaging. The normalization strategy itself is a spec/plan
   decision, not a constitution-level rule.

## Article VI — Simplicity, cost, and local-first
1. The simplest solution that satisfies the spec is preferred (YAGNI).
2. Every component runs in Docker so the whole stack is portable across
   Hetzner, Railway, or Render without code changes.
3. No paid managed service is introduced in the MVP unless a spec justifies
   it.
4. **Local-first development.** The full stack (Go API, PostgreSQL +
   PostGIS, Redis, MinIO, PaddleOCR) MUST run locally via
   `docker compose up` with no cloud accounts and zero cost. Cloud
   deployment (Hetzner + Coolify) is deferred and, when enabled, is a
   configuration change — not a rewrite.

## Article VII — Minimal MVP security
1. Authentication is token-based (JWT). Passwords MUST be hashed with
   bcrypt or argon2 — never stored in plaintext.
2. All endpoints MUST validate input. Receipt upload MUST be rate-limited.
3. Receipt images MUST be stored with unguessable URLs.

## Article VIII — Ready to grow, without over-engineering
1. The architecture MUST allow adding the supermarket API and subscription
   plans in the future WITHOUT rewriting the core — but those pieces are
   NOT built in the MVP.
2. Initial scaling is vertical (raise the VPS plan). Horizontal scaling
   (split workers, read replicas, load balancing) is enabled by the design
   but NOT implemented until metrics demand it.

## Article IX — Working language
1. English is the project convention for all code, identifiers, comments,
   specs, plans, documentation, and commit messages.

## Governance
This constitution prevails over any implementation preference. Any
exception MUST be explicitly documented in the corresponding plan with its
justification and the article it deviates from.

**Amendment procedure.** A change to this constitution requires: (a) a
written proposal citing the article(s) affected, (b) review against
`docs/00_GLOBAL_DESIGN.md` to confirm consistency with the agreed
architecture, and (c) an updated version line (below) plus a Sync Impact
Report prepended as an HTML comment at the top of this file.

**Versioning policy.** Versions follow semantic versioning:
- **MAJOR** — an article is removed or its intent redefined in a
  backward-incompatible way.
- **MINOR** — a new article or materially expanded guidance is added.
- **PATCH** — clarifications, wording, or non-semantic refinements.

**Compliance review.** Every `/speckit.plan` MUST include a Constitution
Check that cites the articles each technical decision satisfies; plans that
cannot cite a supporting article for a non-trivial decision MUST flag it as
a deviation. `/speckit.specify` and `/speckit.tasks` MUST NOT contradict
these articles.

**Maintenance note.** The binding copy is `.specify/memory/constitution.md`.
This file (`docs/03_CONSTITUTION.md`) is a readable mirror and MUST be kept
in sync whenever the binding copy changes (and vice versa).

**Version**: 1.0.0 | **Ratified**: 2026-06-21 | **Last Amended**: 2026-06-21
