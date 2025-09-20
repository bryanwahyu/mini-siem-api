#!/usr/bin/env bash
set -Eeuo pipefail

APP=server-analyst
USER=srv-analyst
OPT_DIR=/opt/${APP}
VAR_DIR=/var/lib/${APP}

if [[ $(id -u) -ne 0 ]]; then
  echo "Please run as root" >&2
  exit 1
fi

id -u "$USER" >/dev/null 2>&1 || useradd --system --home "$OPT_DIR" --shell /usr/sbin/nologin "$USER"
mkdir -p "$OPT_DIR" "$VAR_DIR" "$VAR_DIR/spool" /etc/${APP}
chown -R "$USER":"$USER" "$OPT_DIR" "$VAR_DIR"

# Build if binary not provided
if [[ ! -f ./server-analyst ]]; then
  echo "Building binary..."
  make -C "$(dirname "$0")/.." build
fi

install -m 0755 "$(dirname "$0")/../server-analyst" "$OPT_DIR/${APP}"
ln -sf "$OPT_DIR/${APP}" /usr/local/bin/${APP}

echo "Installed to $OPT_DIR"
