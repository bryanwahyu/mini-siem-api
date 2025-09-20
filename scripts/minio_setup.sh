#!/usr/bin/env bash
set -Eeuo pipefail

: "${MINIO_ENDPOINT:?set MINIO_ENDPOINT}"
: "${MINIO_ACCESS_KEY:?set MINIO_ACCESS_KEY}"
: "${MINIO_SECRET_KEY:?set MINIO_SECRET_KEY}"
: "${MINIO_BUCKET:=server-analyst}"
: "${MINIO_REGION:=us-east-1}"

if ! command -v mc >/dev/null 2>&1; then
  echo "Installing MinIO mc..."
  curl -fsSL -o /usr/local/bin/mc https://dl.min.io/client/mc/release/linux-amd64/mc
  chmod +x /usr/local/bin/mc
fi

mc alias set sa http://"${MINIO_ENDPOINT}" "${MINIO_ACCESS_KEY}" "${MINIO_SECRET_KEY}"
mc mb -p sa/"${MINIO_BUCKET}" || true

cat > /tmp/sa_ilm.json <<'EOF'
{
  "Rules": [
    {"ID":"raw-retention","Status":"Enabled","Filter":{"Prefix":"raw/"},"Expiration":{"Days":30}},
    {"ID":"events-retention","Status":"Enabled","Filter":{"Prefix":"events/"},"Expiration":{"Days":180}},
    {"ID":"detections-retention","Status":"Enabled","Filter":{"Prefix":"detections/"},"Expiration":{"Days":365}},
    {"ID":"decisions-retention","Status":"Enabled","Filter":{"Prefix":"decisions/"},"Expiration":{"Days":365}}
  ]
}
EOF

mc ilm import sa/"${MINIO_BUCKET}" < /tmp/sa_ilm.json
echo "Lifecycle configured. Testing put..."
echo "ok" | mc pipe sa/"${MINIO_BUCKET}"/health/test.txt
mc stat sa/"${MINIO_BUCKET}"/health/test.txt >/dev/null
echo "MinIO setup done."
