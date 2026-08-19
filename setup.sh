#!/bin/bash
# One-command Raspberry Pi installation for MagicBoxie. Run this script from
# the repository root on the Pi. It is safe to re-run for upgrades.
set -euo pipefail

HOSTNAME="${MAGICBOX_HOSTNAME:-magicboxie}"
CONTENT_DIR="${MAGICBOX_CONTENT_DIR:-/content}"
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

if [ ! -f "$SCRIPT_DIR/Makefile" ]; then
  echo "Run setup.sh from a complete MagicBoxie source checkout." >&2
  exit 1
fi

if ! command -v apt-get >/dev/null 2>&1; then
  echo "setup.sh currently supports Raspberry Pi OS/Debian hosts." >&2
  exit 1
fi

case "$(uname -m)" in
  aarch64)
    BUILD_ENV=(CGO_ENABLED=0 GOOS=linux GOARCH=arm64)
    ;;
  armv7l|armv6l)
    BUILD_ENV=(CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7)
    ;;
  *)
    echo "Unsupported Raspberry Pi architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

cd "$SCRIPT_DIR"

echo "==> Installing system packages"
sudo apt-get update
sudo apt-get install -y nginx npm

echo "==> Preparing the Raspberry Pi"
make pi-setup PI_HOSTNAME="$HOSTNAME" PI_CONTENT_DIR="$CONTENT_DIR"

echo "==> Building and installing MagicBoxie"
env "${BUILD_ENV[@]}" make pi-install PI_HOSTNAME="$HOSTNAME" PI_CONTENT_DIR="$CONTENT_DIR"

# nginx owns the public listener, so do not expose the application server's
# port directly on the LAN. Preserve every other user-configured setting.
sudo sed -i 's|^listen_addr:.*|listen_addr: "127.0.0.1:8080"|' /etc/magicbox/config.yaml

echo "==> Installing nginx configuration"
sudo install -m 0644 deploy/pi/nginx/magicboxie.conf /etc/nginx/sites-available/magicboxie
sudo ln -sfn /etc/nginx/sites-available/magicboxie /etc/nginx/sites-enabled/magicboxie
sudo nginx -t

echo "==> Starting services"
sudo systemctl enable --now nginx avahi-daemon
sudo systemctl restart magicbox nginx

echo "Done. Open http://$HOSTNAME.lan or http://$HOSTNAME.local"
