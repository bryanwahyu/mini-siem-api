#!/usr/bin/env bash
set -Eeuo pipefail

: "${MINIO_ENDPOINT:?set MINIO_ENDPOINT}"
: "${MINIO_ACCESS_KEY:?set MINIO_ACCESS_KEY}"
: "${MINIO_SECRET_KEY:?set MINIO_SECRET_KEY}"
: "${MINIO_BUCKET:=server-analyst}"
: "${MINIO_PREFIX:=prod}"
: "${SQLITE_PATH:=/var/lib/server-analyst/app.db}"

TS=$(date +%s)
TMP=$(mktemp -d)
OUT="$TMP/server-analyst-${TS}.db.gz"
mkdir -p "$TMP"
sqlite3 "$SQLITE_PATH" ".backup '$TMP/app.db'"
gzip -c "$TMP/app.db" > "$OUT"

if ! command -v mc >/dev/null 2>&1; then
  curl -fsSL -o /usr/local/bin/mc https://dl.min.io/client/mc/release/linux-amd64/mc
  chmod +x /usr/local/bin/mc
fi
mc alias set sa http://"${MINIO_ENDPOINT}" "${MINIO_ACCESS_KEY}" "${MINIO_SECRET_KEY}"
DATEPATH=$(date +"%Y/%m/%d")
mc cp "$OUT" sa/"${MINIO_BUCKET}"/"${MINIO_PREFIX}"/backups/sqlite/$DATEPATH/server-analyst-${TS}.db.gz

find /var/lib/server-analyst -name "*.db.gz" -type f -mtime +7 -delete || true
rm -rf "$TMP"
echo "Backup uploaded."
