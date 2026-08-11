#!/bin/bash

set -e

INSTALL_DIR="/opt/hkl-agent"
CONFIG_FILE="$INSTALL_DIR/config.json"
SERVICE_FILE="/etc/systemd/system/hkl-agent.service"

echo "===================================="
echo "       HKL-Agent Linux Installer"
echo "===================================="

echo "[1/7] Sprawdzanie uprawnień..."

if [ "$EUID" -ne 0 ]; then
    echo "Uruchom instalator jako root:"
    echo ""
    echo "sudo ./install.sh"
    exit 1
fi

echo "[2/7] Tworzenie katalogu..."

mkdir -p "$INSTALL_DIR"

echo "[3/7] Pobieranie nazwy urządzenia..."

DEVICE_NAME=$(hostname)

if [ -z "$DEVICE_NAME" ]; then
    echo "Nie udało się pobrać nazwy urządzenia."
    exit 1
fi

echo "Wykryto:"
echo "  $DEVICE_NAME"

echo "[4/7] Instalowanie agenta..."

cp HKL-Agent "$INSTALL_DIR/HKL-Agent"
chmod +x "$INSTALL_DIR/HKL-Agent"

echo "[5/7] Tworzenie konfiguracji..."

cat > "$CONFIG_FILE" <<EOF
{
    "device_name": "$DEVICE_NAME",
    "server_url": "http://192.168.1.100:8080",
    "heartbeat_interval": 30
}
EOF

echo "[6/7] Tworzenie usługi systemd..."

cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=HKL Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/HKL-Agent
WorkingDirectory=$INSTALL_DIR
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

echo "[7/7] Uruchamianie agenta..."

systemctl daemon-reload
systemctl enable hkl-agent
systemctl restart hkl-agent

echo ""
echo "===================================="
echo "        INSTALACJA ZAKOŃCZONA"
echo "===================================="
echo ""
echo "Urządzenie: $DEVICE_NAME"
echo "Agent: $INSTALL_DIR/HKL-Agent"
echo "Config: $CONFIG_FILE"
echo ""
echo "Status:"
systemctl status hkl-agent --no-pager
