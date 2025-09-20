#!/usr/bin/env bash
set -Eeuo pipefail

NGINX_DIR=${NGINX_DIR:-/etc/nginx}
CONF_DIR="$NGINX_DIR/server-analyst"
mkdir -p "$CONF_DIR"

cat > "$CONF_DIR/blocklist.conf" <<'EOF'
map $remote_addr $is_blocked_ip {
    default 0;
    # populated dynamically
}

server {
    # include in your server config:
    if ($is_blocked_ip) { return 403; }
}
EOF

nginx -t && systemctl reload nginx || true
echo "Nginx blocklist config prepared at $CONF_DIR/blocklist.conf"
