#!/usr/bin/env bash
set -Eeuo pipefail

if [[ $(id -u) -ne 0 ]]; then
  echo "Please run as root" >&2
  exit 1
fi

if command -v nft >/dev/null 2>&1; then
  echo "Configuring nftables blacklist set..."
  nft list tables inet | grep -q "filter" || nft add table inet filter
  nft list chain inet filter input >/dev/null 2>&1 || nft add chain inet filter input { type filter hook input priority 0 \; }
  nft list set inet filter blacklist >/dev/null 2>&1 || nft add set inet filter blacklist { type ipv4_addr \; flags timeout \; }
  nft list ruleset | grep -q "@blacklist drop" || nft add rule inet filter input ip saddr @blacklist drop
  nft -s list ruleset | sed -n '1,80p'
fi

if command -v ufw >/dev/null 2>&1; then
  echo "UFW present. Ensure enabled..."
  ufw status | grep -q inactive && yes | ufw enable || true
fi

echo "Firewall configured."
