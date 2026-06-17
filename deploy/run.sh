#!/bin/sh
# Container entrypoint: restore-then-replicate.
#
# On a fresh volume (first boot or disaster recovery) the local DB is missing,
# so we restore the latest snapshot from R2. Then we run the API *under*
# Litestream so every write is replicated continuously.
set -e

mkdir -p "$(dirname "$DB_PATH")"

if [ ! -f "$DB_PATH" ]; then
  echo "litestream: no local DB at $DB_PATH; attempting restore from replica…"
  litestream restore -if-replica-exists -config /etc/litestream.yml "$DB_PATH" || true
fi

exec litestream replicate -config /etc/litestream.yml -exec "/app/api"
