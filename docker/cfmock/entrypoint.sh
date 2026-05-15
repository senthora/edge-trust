#!/bin/sh
set -eu

# sets up seed CIDR values
cp -n /seed/ips.json /usr/share/nginx/html/ips.json

# set CIDR values from env vars only if they exist,
# otherwise skip and use seed values
cfmock set

exec nginx -g 'daemon off;'
