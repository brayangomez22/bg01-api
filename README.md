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
make run            # serves http://localhost:8080
curl localhost:8080/health
```

## Make targets

| Target | What |
|---|---|
| `make run` | Run the API server |
| `make build` | Build binary to `bin/api` |
| `make vet` / `make test` | Static analysis / tests |
| `make tools` | Install dev tooling (goose, sqlc) |
| `make migrate-up DB_PATH=bg01.db` | Apply migrations *(Phase 1+)* |
| `make sqlc` | Regenerate DB code *(Phase 1+)* |
