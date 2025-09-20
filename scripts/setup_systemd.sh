#!/usr/bin/env bash
set -Eeuo pipefail

APP=server-analyst
USER=srv-analyst
OPT_DIR=/opt/${APP}
ENV_FILE=/etc/${APP}/${APP}.env
SERVICE=/etc/systemd/system/${APP}.service

if [[ $(id -u) -ne 0 ]]; then
  echo "Please run as root" >&2
  exit 1
fi

cat > "$SERVICE" <<EOF
[Unit]
Description=Server Analyst (threat detector)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${USER}
Group=${USER}
EnvironmentFile=-${ENV_FILE}
ExecStart=${OPT_DIR}/${APP} -config /etc/${APP}/config.yaml
WorkingDirectory=${OPT_DIR}
Restart=on-failure
RestartSec=3s
LimitNOFILE=65536
AmbientCapabilities=CAP_NET_ADMIN
CapabilityBoundingSet=CAP_NET_ADMIN
NoNewPrivileges=true
ProtectSystem=full
ProtectHome=true
PrivateTmp=true
PrivateDevices=true
LogsDirectory=${APP}
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now ${APP}.service
systemctl status ${APP}.service --no-pager -l || true

echo "Systemd unit installed and started."
