#!/bin/sh
set -euo pipefail

APP_USER=notimulti
APP_GROUP=notimulti
APP_UID=$(id -u "$APP_USER")
APP_GID=$(id -g "$APP_GROUP")
APP_IDENTITY="$APP_USER:$APP_GROUP"

CONFIG_DIR=${CONFIG_DIR:-/app/configs}
DB_PATH=${DB_PATH:-/app/data/notifications.db}
DB_DIR=$(dirname "$DB_PATH")

log() {
  printf '[entrypoint] %s\n' "$*"
}

mkdir -p "$CONFIG_DIR" "$DB_DIR"

if chown -R "$APP_UID":"$APP_GID" "$CONFIG_DIR" "$DB_DIR" 2>/tmp/chown.log; then
  log "Ensured ownership of $CONFIG_DIR and $DB_DIR"
else
  log "Warning: could not change ownership of bind mounts (see /tmp/chown.log); continuing"
fi

exec su-exec "$APP_IDENTITY" ./server
