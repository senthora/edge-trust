#!/usr/bin/env bash
set -euo pipefail

mkdir -p ci/actual

compare_config() {
  local container_path="$1"
  local expected_path="$2"
  local actual_path="$3"

  echo "Verifying: ${container_path}"

  docker compose cp \
    "edge-trust:${container_path}" \
    "${actual_path}"

  diff \
    "${expected_path}" \
    <(grep -Ev '^\s*(#|$)' "${actual_path}")

  echo " ✔ Verified: ${container_path}"
  echo
}

compare_state() {
  local container_path="$1"
  local expected_path="$2"
  local actual_path="$3"

  echo "Verifying state: ${container_path}"

  docker compose cp \
    "edge-trust:${container_path}" \
    "${actual_path}"

  diff \
    <(jq -c '.' "${expected_path}") \
    <(jq -c 'del(.written_at)' "${actual_path}")

  echo " ✔ Verified state: ${container_path}"
}

compare_config \
  /etc/nginx/dynamic/trusted-proxy-sources.conf \
  ci/expected/trusted-proxy-sources.conf \
  ci/actual/trusted-proxy-sources.conf

compare_config \
  /etc/nginx/dynamic/origin-allowlist.conf \
  ci/expected/origin-allowlist.conf \
  ci/actual/origin-allowlist.conf

compare_state \
  /var/lib/edge-trust/state.json \
  ci/expected/state.json \
  ci/actual/state.json

docker compose exec -T edge-trust \
  sh -ceu '
    test -f /signal/nginx.reload || {
      echo "ERROR: Missing nginx reload signal: /signal/nginx.reload"
      exit 1
    }
  '

echo " ✔ Verified nginx reload signal"
