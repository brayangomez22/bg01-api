# syntax=docker/dockerfile:1

# ── build ────────────────────────────────────────────────────────────────────
# Pure-Go SQLite (modernc) means CGO_ENABLED=0: a fully static binary, no C
# toolchain, no runtime libs.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ── litestream ───────────────────────────────────────────────────────────────
# Grab the Litestream binary from its official image.
FROM litestream/litestream:0.3.13 AS litestream

# ── runtime ──────────────────────────────────────────────────────────────────
# alpine (not scratch/distroless) because the entrypoint needs a shell to
# restore-then-replicate. ca-certificates lets us reach R2 and the GitHub API;
# su-exec lets run.sh drop from root to the unprivileged app user.
FROM alpine:3.24
RUN apk add --no-cache ca-certificates su-exec \
 && adduser -D -H -u 10001 app
COPY --from=build /out/api /app/api
COPY --from=litestream /usr/local/bin/litestream /usr/local/bin/litestream
COPY deploy/litestream.yml /etc/litestream.yml
COPY deploy/run.sh /app/run.sh
RUN chmod +x /app/run.sh

ENV DB_PATH=/data/bg01.db
EXPOSE 8080
ENTRYPOINT ["/app/run.sh"]
