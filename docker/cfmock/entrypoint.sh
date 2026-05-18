#!/bin/sh
set -eu

cfmock set

exec nginx -g 'daemon off;'
