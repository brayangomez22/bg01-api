#!/bin/sh
# Container entrypoint: restore-then-replicate.
#
# On a fresh volume (first boot or disaster recovery) the local DB is missing,
# so we restore the latest snapshot from R2. Then we run the API *under*
# Litestream so every write is replicated continuously.
set -e

# Fly mounts the volume root-owned, so start as root only long enough to hand
# the data dir to the unprivileged app user, then re-run this script as it.
if [ "$(id -u)" = "0" ]; then
  mkdir -p "$(dirname "$DB_PATH")"
  chown -R app:app "$(dirname "$DB_PATH")"
  exec su-exec app "$0" "$@"
fi

if [ ! -f "$DB_PATH" ]; then
  echo "litestream: no local DB at $DB_PATH; attempting restore from replica…"
  litestream restore -if-replica-exists -config /etc/litestream.yml "$DB_PATH" || true
fi

exec litestream replicate -config /etc/litestream.yml -exec "/app/api"
