#!/bin/bash
# One-time, app-specific setup: creates /var/www/magicboxie.com/config on
# the server and drops configs/magicbox.example.yaml there as a starting
# point, if it isn't already present. Everything else (nginx, certbot,
# Docker, the deploy directory) is handled generically by
# ~/Sites/server-config's bootstrap.sh (run once per host).
#
# config/magicbox.yaml holds secrets (auth.password_hash, tmdb.api_read_token)
# that can't be generated automatically, so this script only ever creates
# it from the example if missing -- it never overwrites an existing one,
# and neither does deploy/deploy.sh. After this runs, SSH in and edit it by
# hand:
#   ssh deploy@104.131.183.186
#   vim /var/www/magicboxie.com/config/magicbox.yaml
#
# To generate the password hash, run it through the image itself once an
# image exists at ghcr.io/kriogenx0/magicbox-web (i.e. after the first CI
# run/push to main):
#   docker run --rm ghcr.io/kriogenx0/magicbox-web hash-password '<your password>'
#
# movies_dir/music_dir/data_dir in the example already point at the
# container-internal paths (/content/movies, /content/music,
# /var/lib/magicbox) that docker-compose.prod.yml's named volumes back --
# leave those as-is; only auth.password_hash and tmdb.api_read_token need
# editing.
set -euo pipefail

SSH_TARGET="deploy@104.131.183.186"
REMOTE_DIR="/var/www/magicboxie.com"
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

echo "==> Ensuring $REMOTE_DIR/config exists"
# Each sudo call takes exactly one path argument, matching the shared
# NOPASSWD grant's "/usr/bin/mkdir -p /var/www/*" / "chown deploy:deploy
# /var/www/*" wildcards one-for-one -- a single call with two paths (e.g.
# `chown deploy:deploy "$REMOTE_DIR" "$REMOTE_DIR/config"`) doesn't match
# the pattern and falls through to an (impossible, non-interactive) sudo
# password prompt instead.
ssh "$SSH_TARGET" "sudo mkdir -p '$REMOTE_DIR' && sudo chown deploy:deploy '$REMOTE_DIR' && sudo mkdir -p '$REMOTE_DIR/config' && sudo chown deploy:deploy '$REMOTE_DIR/config'"

if ssh "$SSH_TARGET" "test -f '$REMOTE_DIR/config/magicbox.yaml'"; then
  echo "==> $REMOTE_DIR/config/magicbox.yaml already exists -- leaving it alone"
else
  echo "==> Seeding config/magicbox.yaml from the example"
  scp "$SCRIPT_DIR/../configs/magicbox.example.yaml" "$SSH_TARGET:$REMOTE_DIR/config/magicbox.yaml"
  echo "==> Edit it now: ssh $SSH_TARGET, then vim $REMOTE_DIR/config/magicbox.yaml"
  echo "    Set auth.password_hash (see this script's header) and tmdb.api_read_token."
fi
