# BG-01 API

Go backend that administers the content of the
[BG-01 Space Station](https://brayangomez.dev) portfolio.

It is the CRUD source of truth for the portfolio's content (pilot, missions,
technologies, experience, training sims, knowledge archive, comms copy). The
public site stays static on GitHub Pages and is rebuilt from this API's
`/export` endpoint at build time — see [CLAUDE.md](./CLAUDE.md) for the full
architecture.

## Stack

Go · stdlib `net/http` · SQLite (`modernc.org/sqlite`) · sqlc · goose

## Quick start

```bash
cp .env.example .env
# Fill in the two required auth values (the API refuses to start without them):
#   ADMIN_PASSWORD_HASH  ← go run ./cmd/hashpw 'your-password'
#   SESSION_SECRET       ← openssl rand -base64 32   (≥ 32 bytes)
# For local HTTP also set COOKIE_SECURE=false.
make run            # serves http://localhost:8080
curl localhost:8080/health
```

## Security notes

- Startup fails unless `ADMIN_PASSWORD_HASH` is set and `SESSION_SECRET` is at
  least 32 bytes (an empty secret would make session tokens forgeable).
- `POST /admin/login` is rate limited to 5 attempts per minute per client IP
  (`Fly-Client-IP`, falling back to the peer address) and answers `429` with
  `Retry-After`.
- The public `POST /pulse` visit beacon is limited to 10 hits per minute per IP
  (stops fake-visitor floods); `X-Forwarded-For` is ignored.
- Content upserts require a URL-safe primary key
  (`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`), otherwise `400`.
- Request bodies are capped at 1 MB (`413`); the server sets read, write and
  idle timeouts.

## Make targets

| Target | What |
|---|---|
| `make run` | Run the API server |
| `make build` | Build binary to `bin/api` |
| `make vet` / `make test` | Static analysis / tests (in-memory SQLite, no setup) |
| `make tools` | Install dev tooling (goose, sqlc) |
| `make migrate-up DB_PATH=bg01.db` | Apply migrations *(Phase 1+)* |
| `make sqlc` | Regenerate DB code *(Phase 1+)* |
