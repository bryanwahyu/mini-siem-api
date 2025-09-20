#!/usr/bin/env bash
set -Eeuo pipefail

APP=server-analyst
USER=srv-analyst

if [[ $(id -u) -ne 0 ]]; then
  echo "Please run as root" >&2
  exit 1
fi

systemctl disable --now ${APP}.service || true
rm -f /etc/systemd/system/${APP}.service
systemctl daemon-reload

rm -f /usr/local/bin/${APP}
rm -rf /opt/${APP}

id -u "$USER" >/dev/null 2>&1 && userdel -r "$USER" || true

echo "Uninstalled ${APP}."
