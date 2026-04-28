#!/usr/bin/env bash
set -euo pipefail

export FILEBOT_HOME="${FILEBOT_HOME:-/config/filebot}"

init_filebot_license() {
  local lpath="/config/filebot/license.psm"
  mkdir -p /config/filebot/data >/dev/null 2>&1 || true

  if [ -d /opt/filebot ]; then
    rm -rf /opt/filebot/data
    ln -sf /config/filebot/data /opt/filebot/data
  fi

  if [ -x /usr/local/bin/filebot ] && [ -f "$lpath" ]; then
    if [ ! -f "/config/filebot/data/.license" ]; then
      echo "[entrypoint] filebot: activating license from $lpath"
      /usr/local/bin/filebot --license "$lpath" >/tmp/filebot-license.log 2>&1 || true
    else
      echo "[entrypoint] filebot: license already activated in persistent storage"
    fi
  fi
}

init_filebot_license
exec /app/winston
