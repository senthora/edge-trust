#!/bin/sh
set -eu

mkdir -p /test/expected

ALL_CIDRS="$(printf '%s,%s' "${IPV4_CIDRS}" "${IPV6_CIDRS}" | tr ',' ' ')"

generate_trusted_proxies() {
  echo "Generating trusted proxy sources"

  {
    for cidr in ${ALL_CIDRS}; do
      echo "set_real_ip_from ${cidr};"
    done
  } > /test/expected/trusted-proxy-sources.conf

  echo " ✔ Generated trusted proxy sources"
  echo
}

generate_origin_allowlist() {
  echo "Generating origin allowlist"

  {
    for cidr in ${ALL_CIDRS}; do
      echo "${cidr} 1;"
    done
  } > /test/expected/origin-allowlist.conf

  echo " ✔ Generated origin allowlist"
  echo
}

generate_state() {
  echo "Generating expected state"

  jq -n \
    --arg source_url "http://cfmock/client/v4/ips" \
    --arg etag "${ETAG}" \
    --arg hash "${HASH}" \
    --argjson cidrs "$(printf '%s\n' ${ALL_CIDRS} | jq -R . | jq -s .)" \
    '{
      source_url: $source_url,
      etag: $etag,
      cidrs: $cidrs,
      hash: $hash
    }' \
    > /test/expected/state.json

  echo " ✔ Generated expected state"
  echo
}

assert_config() {
  expected_path="$1"
  actual_path="$2"

  echo "Comparing: ${actual_path}"

  grep -Ev '^\s*(#|$)' "${actual_path}" \
    > /tmp/actual-config

  diff \
    "${expected_path}" \
    /tmp/actual-config

  echo " ✔ Verified"
  echo
}

assert_state() {
  expected_path="$1"
  actual_path="$2"

  echo "Comparing state: ${actual_path}"

  jq -c '.' "${expected_path}" \
    > /tmp/expected-state

  jq -c 'del(.written_at)' "${actual_path}" \
    > /tmp/actual-state

  diff \
    /tmp/expected-state \
    /tmp/actual-state

  echo " ✔ Verified"
  echo
}

assert_signal() {
  actual_path="$1"

  echo "Comparing signal: ${actual_path}"

  if [ ! -f "${actual_path}" ]; then
    echo "ERROR: Missing signal file: ${actual_path}"
    exit 1
  fi

  echo " ✔ Verified"
  echo
}

generate_trusted_proxies
generate_origin_allowlist
generate_state

assert_config \
  /test/expected/trusted-proxy-sources.conf \
  /dynamic/trusted-proxy-sources.conf

assert_config \
  /test/expected/origin-allowlist.conf \
  /dynamic/origin-allowlist.conf

assert_state \
  /test/expected/state.json \
  /state/state.json

assert_signal \
  /signal/nginx.reload

echo "Integration tests passed"
