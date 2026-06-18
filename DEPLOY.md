# Deploy — BG-01 API

The API runs as a single small container on **Fly.io** with a persistent volume
for the SQLite database, and **Litestream** continuously replicating that
database to **Cloudflare R2** for offsite backup / disaster recovery.

Artifacts: [`Dockerfile`](./Dockerfile), [`fly.toml`](./fly.toml),
[`deploy/litestream.yml`](./deploy/litestream.yml),
[`deploy/run.sh`](./deploy/run.sh).

> The image is verified locally (`docker build`); the steps below are the
> interactive ones that need your accounts. Run them yourself (e.g. with the
> `!` prefix in the session, or your own terminal).

---

## 1. Prerequisites

```sh
# flyctl (Arch: official package or the installer)
curl -L https://fly.io/install.sh | sh
fly auth login
```

## 2. Cloudflare R2 (Litestream target)

1. Cloudflare dashboard → **R2** → create bucket `bg01-litestream`.
2. **R2 → Manage API Tokens → Create API Token** (Object Read & Write, scoped to
   the bucket). Note the **Access Key ID** and **Secret Access Key**.
3. Your S3 endpoint is `https://<ACCOUNT_ID>.r2.cloudflarestorage.com`
   (Account ID is on the R2 overview page).

## 3. Create the Fly app + volume

```sh
cd bg01-api
fly apps create bg01-api                      # name must match fly.toml
fly volumes create bg01_data --region iad --size 1   # 1 GB is ample; region must match fly.toml
```

## 4. Secrets

Generate fresh production credentials (do not reuse local dev values):

```sh
# Admin password hash (pick a strong password):
go run ./cmd/hashpw 'YOUR-PROD-PASSWORD'      # copy the $2a$… output

fly secrets set \
  ADMIN_PASSWORD_HASH='<hash from above>' \
  SESSION_SECRET="$(openssl rand -base64 32)" \
  GITHUB_DISPATCH_TOKEN='<github PAT, see step 7>' \
  LITESTREAM_ENDPOINT='https://<ACCOUNT_ID>.r2.cloudflarestorage.com' \
  LITESTREAM_ACCESS_KEY_ID='<R2 access key id>' \
  LITESTREAM_SECRET_ACCESS_KEY='<R2 secret access key>'
```

Non-secret config (origin, DB path, bucket, publish repo) is already in
`fly.toml` under `[env]`.

## 5. Deploy

```sh
fly deploy
fly logs            # watch: migrations run, "station online", Litestream replicating
curl https://bg01-api.fly.dev/health   # {"status":"ok"}
```

## 6. Custom domain `api.brayangomez.dev`

```sh
fly certs add api.brayangomez.dev
fly ips list        # note the v4 (shared/dedicated) and v6
```

In Cloudflare DNS for `brayangomez.dev`, add:

- `CNAME  api  bg01-api.fly.dev`  — **DNS only (grey cloud)**, not proxied, so
  the session cookie stays same-site and Fly terminates TLS.

Wait for `fly certs show api.brayangomez.dev` to report the cert as issued.

## 7. GitHub token for the Publish button

Create a **fine-grained PAT** (github.com → Settings → Developer settings):
- Repository access: only `brayangomez22/bg01-station`.
- Permissions: **Contents: Read and write** (enough to trigger
  `repository_dispatch`).

Use it as `GITHUB_DISPATCH_TOKEN` in step 4. After this, the "Publicar" button
in `/control` fires the deploy workflow.

## 8. Seed production (first deploy only)

The fresh production DB is empty. Seed it once from the committed snapshot:

```sh
# from the portfolio repo (bg01-station)
API_BASE=https://api.brayangomez.dev SEED_PASSWORD='YOUR-PROD-PASSWORD' npm run seed
```

Litestream will replicate the seeded DB to R2 automatically. (On any future
volume loss, `run.sh` restores from R2 on boot.)

## 9. Point the portfolio at the API

In `bg01-station`:
- GitHub repo → **Settings → Secrets and variables → Actions → Variables**: add
  `API_BASE = https://api.brayangomez.dev` (the deploy workflow reads it).
- For local builds that should hit prod, set `VITE_API_BASE_URL` in `.env`.
  (Dev defaults to `http://localhost:8080`.)

---

## Operations

- **Logs:** `fly logs`
- **SSH:** `fly ssh console` → the DB is at `/data/bg01.db`.
- **Manual backup check:** `fly ssh console -C "litestream snapshots -config /etc/litestream.yml /data/bg01.db"`
- **Restore drill:** delete the volume and redeploy; `run.sh` restores from R2.
- **Cost:** machines auto-stop when idle (`min_machines_running = 0`) and wake on
  the next request, so the API costs ~nothing between edits.
