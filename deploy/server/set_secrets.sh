#!/bin/bash
# Run this ON THE SERVER (unlike this repo's other deploy/*.sh, which run
# from your machine and SSH out) -- it prompts for the login password and
# TMDB token and writes them into config/magicbox.yaml in place, leaving
# every other setting (movies_dir, transcode, etc.) untouched. Safe to
# re-run any time to rotate either value.
#
#   scp deploy/server/set_secrets.sh deploy@104.131.183.186:~
#   ssh -t deploy@104.131.183.186 ./set_secrets.sh
#
# No sudo needed -- deploy already owns /var/www/magicboxie.com/config
# (server_setup.sh's job). The password hash is computed with a
# `docker run --rm httpd:2.4-alpine htpasswd -nbBC 10 ...` one-liner rather
# than `docker run ghcr.io/kriogenx0/magicbox-web hash-password ...`
# (config/magicbox.yaml's own comment, and server_setup.sh's header) so
# this doesn't depend on that image existing in GHCR yet -- htpasswd -B is
# plain bcrypt, cost 10, same as the app's own
# `bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)`
# (internal/auth.go) -- the $2y$ vs $2a$ prefix difference is cosmetic;
# Go's bcrypt.CompareHashAndPassword accepts both.
set -euo pipefail

CONFIG="/var/www/magicboxie.com/config/magicbox.yaml"

test -f "$CONFIG" || {
  echo "Missing $CONFIG -- run deploy/server_setup.sh from your machine first." >&2
  exit 1
}

read -r -s -p "New MagicBox login password: " PASSWORD
echo
read -r -s -p "Confirm password: " PASSWORD_CONFIRM
echo
[ "$PASSWORD" = "$PASSWORD_CONFIRM" ] || { echo "Passwords didn't match." >&2; exit 1; }
[ -n "$PASSWORD" ] || { echo "Password can't be empty." >&2; exit 1; }

read -r -p "TMDB v4 read access token (blank to leave unchanged): " TMDB_TOKEN

echo "==> Hashing password"
HASH=$(docker run --rm httpd:2.4-alpine htpasswd -nbBC 10 x "$PASSWORD" | cut -d: -f2)

echo "==> Writing $CONFIG"
TMP=$(mktemp)
awk -v hash="$HASH" -v token="$TMDB_TOKEN" -v set_token="$([ -n "$TMDB_TOKEN" ] && echo 1 || echo 0)" '
  /^[[:space:]]*password_hash:/ { print "  password_hash: \"" hash "\""; next }
  /^[[:space:]]*api_read_token:/ {
    if (set_token == "1") { print "  api_read_token: \"" token "\""; next }
  }
  { print }
' "$CONFIG" > "$TMP"
mv "$TMP" "$CONFIG"
chmod 600 "$CONFIG"

echo "==> Done. Restart the container to pick this up, if it's already running:"
echo "    cd /var/www/magicboxie.com && docker compose restart magicbox"
