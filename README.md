# TraumaCompendium — Backend

> Learn. Heal. Connect. Stay Safe.

Go backend for **TraumaCompendium**, a trauma education & recovery platform.
This repository implements the **Phase 1 core foundation** of the MVP:

- **Auth & profiles** — registration, email + phone OTP verification, login,
  JWT access/refresh sessions with rotation, forgot/reset password.
- **Structured learning (Beginner level)** — levels, lessons, quizzes,
  per-user progress, and badges.
- **Journaling** — private entries with JSON/CSV export.
- **Mood & trigger tracking** — quick daily logs with a trend summary.
- **Localization** — learning content (levels, lessons, quizzes, badges) and
  validation messages in English plus 7 other languages, with a
  text-to-speech lesson-audio endpoint. See [Localization](#localization)
  below — the translations are LLM-drafted and explicitly unreviewed.

SafeReach emergency alerts, the AI assistant, community features, levels 2–6
content, and paid tiers are Phase 2/3 scope per the product brief and are not
in this repository yet.

> ⚠️ This platform does not provide medical diagnosis, therapy, or emergency
> dispatch. It is an educational and self-tracking tool only.

## Tech stack

| Layer          | Choice                                                         |
|-----------------|------------------------------------------------------------------|
| Language        | Go 1.26                                                          |
| HTTP framework  | [Gin](https://github.com/gin-gonic/gin)                          |
| Database        | MySQL 8, via `database/sql` + `go-sql-driver/mysql` (no ORM)     |
| Migrations      | Hand-rolled, embedded `.sql` files (see `internal/platform/database`) |
| Cache / limits  | Redis (`redis/go-redis/v9`) — rate limiting, refresh-token support |
| Auth            | JWT (`golang-jwt/v5`), bcrypt passwords, rotating refresh tokens |
| SMS             | [Termii](https://termii.com), sandbox/production switch         |
| Email           | SMTP (stdlib `net/smtp`), sandbox/production switch              |
| Text-to-speech  | Google Cloud TTS REST API, sandbox/production switch              |
| Logging         | `zerolog`, structured JSON (console-pretty in dev)                |
| Validation      | `go-playground/validator` via Gin binding                        |

**Why no ORM, no GORM, no golang-migrate?** This project was built in an
offline environment with a pre-populated Go module cache; those specific
libraries weren't available in it. The result — hand-written SQL behind
repository interfaces, and a small embedded migrator — is a perfectly
standard, arguably more transparent way to build a Go backend, so it stayed
that way rather than being swapped out once network access was available
again. See `internal/platform/database/migrator.go` for the migrator.

## Architecture

Layered, dependency-inverted, one module per domain:

```
cmd/api/main.go           entrypoint: wiring + graceful shutdown

internal/
  config/                  typed env config, fails fast on missing secrets
  domain/                  entities + repository interfaces (no I/O)
    user/ auth/ journal/ mood/ learning/
  repository/mysql/        database/sql implementations of those interfaces
  service/                 business logic, depends only on domain interfaces
    authservice/ userservice/ journalservice/ moodservice/ learningservice/
  http/
    handler/                Gin handlers: decode -> call one service method -> respond
    middleware/              auth, RBAC, rate limiting, logging, recovery, CORS
    dto/                     wire request/response shapes, decoupled from domain types
    router/                  route wiring
  notification/             email + SMS + TTS senders (sandbox/production per provider)
  platform/                 database, cache, logger bootstrapping
  pkg/                      apperror, jwtutil, hash, otp, i18n — small, dependency-free utilities
  testsupport/fakes/        in-memory repository/sender fakes shared by service & handler tests

migrations/                 versioned, embedded .sql files (golang-migrate-style naming)
docs/openapi.yaml           full API contract
```

**Request flow:** `router` → `middleware` (auth/rate-limit) → `handler` (bind +
validate) → `service` (business logic, returns `apperror.Error` on failure) →
`repository` (SQL) → MySQL. `apperror.Error` carries its own HTTP status and
machine-readable `code`, so `http.Fail` maps any service error to a response
without per-handler `switch` statements, and an *unexpected* error is always
reduced to an opaque 500 — internal detail never reaches a client.

**Ownership rule:** every journal/mood/learning-progress query is scoped by
`user_id` at the SQL level, not just checked after the fact — reading or
writing another user's resource by ID returns the same "not found" a
genuinely missing resource would.

## Getting started

### Option A — Docker Compose (recommended)

```sh
docker compose up --build
```

Brings up the API (`:8080`), MySQL 8, and Redis, with sandbox SMS/email
(nothing is sent to a real provider) and migrations applied automatically on
boot. Check it's alive:

```sh
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

### Option B — Local Go + your own MySQL/Redis

```sh
cp .env.example .env
# edit .env: at minimum set JWT_ACCESS_SECRET / JWT_REFRESH_SECRET to two
# different random strings, and point DB_*/REDIS_ADDR at your instances
go run ./cmd/api
```

The app runs its own migrations on startup — no separate migrate step
needed in either option.

## Configuration

All configuration is environment variables (see `.env.example` for the full,
documented list). Highlights:

- `TERMII_MODE` / `SMTP_MODE` / `TTS_MODE` — `sandbox` (default; logs what
  would have been sent, or for TTS returns a placeholder clip — no real
  SMS/email/synthesis) or `production` (real Termii/SMTP/Google Cloud TTS
  calls, requires `TERMII_API_KEY` / `SMTP_HOST` + `SMTP_USERNAME` /
  `TTS_API_KEY` respectively).
- `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` — required, must differ from
  each other. Generate with `openssl rand -base64 48`.
- In `APP_ENV=production`, the app refuses to start if `DB_PASSWORD` is
  empty or if a `_MODE=production` integration is missing its credentials —
  see `internal/config/config.go`'s `validate()`.

## API documentation

The full contract is in [`docs/openapi.yaml`](docs/openapi.yaml) — paste it
into [editor.swagger.io](https://editor.swagger.io) or any OpenAPI viewer.

Every response uses one JSON envelope:

```json
{ "data": { ... }, "meta": { ... } }
{ "error": { "code": "not_found", "message": "...", "fields": { ... } } }
```

`meta` appears only on paginated list endpoints; `fields` only on
`validation_error`.

## Localization

Supported locales: `en` (canonical), `pcm` (Nigerian Pidgin), `yo`
(Yoruba), `ha` (Hausa), `ig` (Igbo), `fr` (French), `sw` (Swahili), `ar`
(Arabic) — see `learning.SupportedLocales`.

- **Requesting a locale**: pass `?locale=yo` on any `/api/v1/learning/*`
  request, or set the `Accept-Language` header; an unrecognized or absent
  locale falls back to English rather than erroring.
- **`needs_review`**: every translatable response (levels, lessons, quiz
  questions, badges) includes `"locale"` and `"needs_review"`. All 210
  seeded non-English rows (`migrations/0021_seed_level1_translations.up.sql`)
  are LLM-drafted and marked `needs_review: true` — **they have not been
  reviewed by a native-speaking translator or a clinician**, which matters
  for trauma-sensitive content. Treat them as a working draft for a human
  reviewer, not as copy ready to ship to end users. Canonical English rows
  are always `needs_review: false`.
- **Lesson audio**: `GET /api/v1/learning/lessons/:id/audio?locale=yo`
  synthesizes (and caches) a lesson's text as speech on first request;
  later requests for the same lesson+locale are served from the
  `lesson_audio` cache. In `TTS_MODE=sandbox` (the default) this returns a
  valid but silent placeholder WAV — real audio requires a configured
  `TTS_API_KEY`. Which of the 8 locales Google Cloud TTS actually
  supports with real voices hasn't been verified against the live API
  from this environment; an unsupported locale fails clearly rather than
  silently (see `internal/notification/tts`).
- **Validation messages** (`required`, `email`, `min`, `max`, `e164`, and
  a generic fallback) are also localized, via `internal/pkg/i18n`'s small
  catalog — independent of, and much narrower than, the content
  translations above.

## Testing

```sh
make test          # full suite (integration tests skip without a real DB/Redis)
make test-unit      # unit tests only, no external services needed
golangci-lint run ./...
```

Repository-layer tests run against a **real** MySQL instance and
rate-limit middleware tests against a **real** Redis instance — both are
skipped automatically (not faked) if `TEST_DATABASE_DSN` / `REDIS_ADDR`
aren't reachable, so `go test ./...` still passes with nothing running
locally. Point them at a throwaway database to run them for real:

```sh
export TEST_DATABASE_DSN="user:pass@tcp(127.0.0.1:3306)/trauma_compendium_test?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC"
export REDIS_ADDR="127.0.0.1:6379"
go test ./...
```

Service-layer tests run against in-memory fakes
(`internal/testsupport/fakes`), so they need nothing at all. CI
(`.github/workflows/ci.yml`) runs everything — vet, build, lint, and the full
test suite — against real MySQL/Redis service containers on every push and
PR.

## Project conventions worth knowing

- **Sandbox-by-default integrations.** `TERMII_MODE`, `SMTP_MODE`, and
  `TTS_MODE` all default to `sandbox`; a fresh checkout, `docker compose
  up`, and CI never talk to a real SMS/email/TTS provider unless someone
  deliberately flips the switch and supplies real credentials.
- **Timestamps are application-owned.** `created_at`/`updated_at` are set
  by the service layer (`Service.now`, swappable in tests) and written
  explicitly, rather than left to the database's `DEFAULT
  CURRENT_TIMESTAMP` — this keeps the in-memory struct a caller already
  holds accurate without a read-back, and makes time-dependent logic (OTP
  expiry, resend cooldowns) deterministically testable.
- **No quiz answer leakage.** `learning.Option.IsCorrect` is scored
  server-side only; the DTO returned by `GET .../quiz` is a distinct type
  that structurally cannot carry it.
- **Refresh token rotation.** Every `/auth/refresh` call revokes the
  presented token and issues a new one — replaying an old, already-rotated
  refresh token fails, which is what makes rotation useful as theft
  detection rather than just bookkeeping.

## What's intentionally not here yet

Scoped out of Phase 1 by design, not by oversight:

- SafeReach emergency alerts, community/peer support, the AI assistant,
  levels 2–6 content, and subscription billing (Phase 2/3 per the brief).
- PDF certificate generation — badges (already implemented) are the Phase 1
  stand-in for "Quizzes, badges, certificates."
- An admin content-management API — levels/lessons/quizzes are authored via
  migration (`migrations/0014_seed_levels_and_beginner_content.up.sql`), not
  an in-app CMS, since Phase 1 has exactly one piece of content to manage.
- Quiz submission history — `SubmitQuiz` scores on the fly and intentionally
  doesn't persist attempts (see the doc comment on
  `learningservice.Service.SubmitQuiz`).
- Localizing `apperror`'s ~60 hand-written service-level messages
  ("invalid credentials", "could not create account", etc.) — only the
  small validator-message catalog is localized (see
  [Localization](#localization)). Doing the rest properly means
  `apperror.Error` carrying message *keys* instead of literal strings and
  touching nearly every service file; a real architectural change, not a
  catalog addition.
- Human review of the 210 seeded non-English content rows, and of the
  `tts.localeVoice` BCP-47 mapping against what Google Cloud TTS actually
  supports live — both require resources (a native-speaking translator /
  clinician, and outbound internet access to the real TTS API) this
  environment didn't have.
