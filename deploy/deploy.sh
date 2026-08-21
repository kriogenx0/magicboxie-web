#!/bin/bash
# Deploys magicboxie-web to production end to end: copies the compose file,
# picks (or reuses) a stable HOST_PORT, brings up the magicboxie container,
# and installs this app's own nginx vhost + TLS cert. Run from CI
# (.github/workflows/deploy.yml) or by hand -- either way it just
# SSHes/SCPs to deploy@104.131.183.186; no other tooling required. Safe to
# re-run.
#
# ~/Sites/server-config is bootstrapping-only (installs Docker/nginx/
# certbot, creates the deploy account + its shared sudoers grant, ships the
# nginx snippets this vhost includes) -- this script owns everything
# specific to *this* app. See that repo's README "Conventions for per-app
# deploy scripts".
#
# Needs, on top of the shared sudoers grant bootstrap.sh installs: the
# per-app grant at deploy/server/sudoers.d/magicboxie-web (installed once, by
# hand, as admin -- see that file's own header), and config/magicboxie.yaml
# already present on the server (deploy/server_setup.sh creates it from
# the example -- see that script; this deploy script never writes or
# overwrites it, since it holds the auth password hash and TMDB token).
set -euo pipefail

SSH_TARGET="deploy@104.131.183.186"
DOMAIN="magicboxie.com"
REMOTE_DIR="/var/www/$DOMAIN"
CERTBOT_EMAIL="simplex0@gmail.com"
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
CERTBOT_CMD="sudo certbot certonly --webroot -w /var/www/certbot -d $DOMAIN -d www.$DOMAIN --non-interactive --agree-tos -m $CERTBOT_EMAIL --deploy-hook 'systemctl reload nginx'"

echo "==> Ensuring $REMOTE_DIR exists"
ssh "$SSH_TARGET" "sudo mkdir -p '$REMOTE_DIR' && sudo chown deploy:deploy '$REMOTE_DIR'"

echo "==> Copying compose file"
scp "$SCRIPT_DIR/../docker-compose.prod.yml" "$SSH_TARGET:$REMOTE_DIR/docker-compose.yml"

echo "==> Checking config/magicboxie.yaml exists (deploy/server_setup.sh creates it -- this script never writes it)"
ssh "$SSH_TARGET" "test -f '$REMOTE_DIR/config/magicboxie.yaml'" || {
  echo "Missing $REMOTE_DIR/config/magicboxie.yaml on the server." >&2
  echo "Run deploy/server_setup.sh once first (see that script's header)." >&2
  exit 1
}

echo "==> Picking (or reusing) a host port"
HOST_PORT=$(ssh "$SSH_TARGET" bash -s -- "$REMOTE_DIR" <<'REMOTE'
set -e
cd "$1"
port=$(grep -m1 '^HOST_PORT=' .env 2>/dev/null | cut -d= -f2 || true)
if [ -z "$port" ]; then
  port=$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1",0)); print(s.getsockname()[1]); s.close()')
fi
grep -q '^HOST_PORT=' .env 2>/dev/null && sed -i "s/^HOST_PORT=.*/HOST_PORT=$port/" .env || echo "HOST_PORT=$port" >> .env
echo "$port"
REMOTE
)
echo "    HOST_PORT=$HOST_PORT"

echo "==> Pulling + recreating the magicboxie container"
ssh "$SSH_TARGET" "cd '$REMOTE_DIR' && docker compose pull magicboxie && docker compose up -d --no-deps magicboxie && docker image prune -f"

install_vhost() {
  local SRC="$1"
  local RENDERED; RENDERED=$(mktemp)
  sed "s/__HOST_PORT__/$HOST_PORT/g" "$SRC" > "$RENDERED"
  scp "$RENDERED" "$SSH_TARGET:/tmp/$(basename "$SRC")"
  rm -f "$RENDERED"
  ssh "$SSH_TARGET" "sudo cp /tmp/$(basename "$SRC") /etc/nginx/sites-available/$DOMAIN && sudo ln -sf /etc/nginx/sites-available/$DOMAIN /etc/nginx/sites-enabled/$DOMAIN && sudo nginx -t && sudo systemctl reload nginx"
}

if ssh "$SSH_TARGET" "sudo test -f /etc/letsencrypt/live/$DOMAIN/cert.pem"; then
  echo "==> Cert already exists"
else
  echo "==> No cert yet — installing the HTTP-only bootstrap vhost first"
  install_vhost "$SCRIPT_DIR/server/sites-available/$DOMAIN.bootstrap.conf"

  echo "==> Requesting the cert"
  ssh "$SSH_TARGET" "$CERTBOT_CMD"
fi

echo "==> Installing the full (HTTPS) vhost"
install_vhost "$SCRIPT_DIR/server/sites-available/$DOMAIN.conf"

echo "==> Renewing the cert if due (no-op otherwise — webroot mode never touches the vhost)"
ssh "$SSH_TARGET" "$CERTBOT_CMD"

echo "Done. https://$DOMAIN -> 127.0.0.1:$HOST_PORT"
