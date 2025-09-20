#!/usr/bin/env bash
set -Eeuo pipefail

if [[ $(id -u) -ne 0 ]]; then
  echo "Please run as root" >&2
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y --no-install-recommends \
  nftables ufw jq curl sqlite3 git unzip make ca-certificates \
  systemd-sysv

if ! command -v go >/dev/null 2>&1; then
  echo "Go not found. Installing Go 1.22.x ..."
  ARCH=$(dpkg --print-architecture)
  case "$ARCH" in
    amd64) GOARCH=amd64 ;;
    arm64) GOARCH=arm64 ;;
    *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;;
  esac
  TMP=$(mktemp -d)
  cd "$TMP"
  curl -fsSL -o go.tar.gz https://go.dev/dl/go1.22.6.linux-${GOARCH}.tar.gz
  rm -rf /usr/local/go
  tar -C /usr/local -xzf go.tar.gz
  ln -sf /usr/local/go/bin/go /usr/local/bin/go
  cd - >/dev/null
fi

echo "Dependencies installed."
