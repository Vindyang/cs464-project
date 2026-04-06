#!/bin/sh
set -eu

mkdir -p /app/data /run/nginx /var/log/supervisor

exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf