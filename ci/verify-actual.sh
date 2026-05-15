#!/usr/bin/env bash
set -euo pipefail

mkdir -p tmp/ci/actual

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
  echo
}

compare_signal() {
  local container_path="$1"
  local actual_path="$2"

  echo "Verifying signal: ${container_path}"

  if ! docker compose cp \
    "edge-trust:${container_path}" \
    "${actual_path}" >/dev/null 2>&1; then
    echo "ERROR: Missing signal file: ${container_path}"
    exit 1
  fi

  echo " ✔ Verified signal: ${container_path}"
  echo
}

compare_config \
  /etc/nginx/dynamic/trusted-proxy-sources.conf \
  tmp/ci/expected/trusted-proxy-sources.conf \
  tmp/ci/actual/trusted-proxy-sources.conf

compare_config \
  /etc/nginx/dynamic/origin-allowlist.conf \
  tmp/ci/expected/origin-allowlist.conf \
  tmp/ci/actual/origin-allowlist.conf

compare_state \
  /var/lib/edge-trust/state.json \
  tmp/ci/expected/state.json \
  tmp/ci/actual/state.json

compare_signal \
  /signal/nginx.reload \
  tmp/ci/actual/nginx.reload
