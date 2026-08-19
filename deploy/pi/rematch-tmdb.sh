#!/bin/bash
set -euo pipefail

sudo -u magicbox /opt/magicbox/bin/magicbox rematch-tmdb --config /etc/magicbox/config.yaml
