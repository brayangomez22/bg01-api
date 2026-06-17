# CLAUDE.md — BG-01 API

Backend that administers all content for the **BG-01 Space Station** portfolio
(the frontend lives in a separate repo at `/home/bra/Projects/portfolio`, deployed
to GitHub Pages). This service is the CRUD source of truth for that content.

## Architecture: build-time export (decided deliberately)

The public portfolio is and stays **static on GitHub Pages**. It does **not** fetch
this API at runtime. Instead:

```
Admin panel ──CRUD──▶ Go API ──▶ SQLite (single file)
                         │ on save, triggers…
                         ▼
   GitHub Action (in the portfolio repo): GET /export → write src/content/*.json
   → npm run build → deploy to Pages
```

So the two repos are decoupled and communicate over an **HTTP JSON contract**, not
shared code. The frontend never imports Go. The API only needs to be awake when
Brayan edits content; otherwise it can sleep (cheap/free hosting). Rationale: keep
Lighthouse/performance intact, keep hosting near-free, while still being a real,
demonstrable Go backend (a showcase repo in its own right).

The **export contract** mirrors the frontend's `src/types/domain.ts`: Pilot,
Mission, Technology, Experience, TrainingSim, ArchiveRecord/ArchiveMeta, Frequency,
plus editable site copy. When the frontend's domain types change, the export shape
must follow.

## Stack

- **Router:** stdlib `net/http` (Go 1.22+ method-aware `ServeMux`). chi only if/when middleware ergonomics demand it.
- **DB:** SQLite via `modernc.org/sqlite` (pure Go, no CGo). Single admin user → SQLite is the right call, not Postgres.
- **Queries:** sqlc (type-safe Go from SQL).
- **Migrations:** goose (`migrations/`).
- **Auth:** single admin — bcrypt password + signed session cookie (Phase 2).
- **Backups:** Litestream (Phase 5).
- **Config:** env vars, `config.Load()` with defaults. See `.env.example`.

Pragmatic schema: list fields (highlights, challenges, metrics, technologies refs,
gallery, archive body segments, refs) are stored as **JSON columns**, not child
tables — it's a single-user CMS, normalization would be busywork. The
mission↔technology cross-index (`usedInMissions`) is **computed in the export**,
mirroring how the frontend computes it today.

## Layout

```
cmd/api/           entrypoint (graceful shutdown)
internal/config/   env config
internal/server/   routing, middleware, handlers
internal/store/    DB layer (sqlc output lands here)  [Phase 1]
internal/auth/     admin auth                          [Phase 2]
migrations/        goose SQL migrations                [Phase 1]
queries/           sqlc query files                    [Phase 1]
```

## Commands

- `make run` — run the server (default `:8080`, `GET /health`)
- `make build` / `make vet` / `make test`
- `make tools` — install goose + sqlc
- `make migrate-up DB_PATH=bg01.db` / `make sqlc` — Phase 1+

## Roadmap

- Phase 0 — scaffold (✅ module, config, server, /health, graceful shutdown)
- Phase 1 — ✅ schema + migrations + sqlc (domain.ts → SQLite)
- Phase 2 — ✅ CRUD API + admin auth + public `/export`
- Phase 3 — ✅ admin panel UI (lives in the portfolio repo, /control)
- Phase 4 — ✅ build-time export wired into the portfolio's GitHub Action;
  `POST /admin/publish` fires a repository_dispatch
- Phase 5 — deploy artifacts ready (Dockerfile, fly.toml, deploy/litestream.yml +
  run.sh, DEPLOY.md); image verified locally. Pending: run `fly deploy`, set
  secrets, R2 bucket, and DNS for api.brayangomez.dev — see DEPLOY.md.
