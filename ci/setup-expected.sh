#!/usr/bin/env bash
set -euo pipefail

mkdir -p ci/expected

IFS=',' read -ra IPV4 <<< "${IPV4_CIDRS}"
IFS=',' read -ra IPV6 <<< "${IPV6_CIDRS}"

ALL_CIDRS=("${IPV4[@]}" "${IPV6[@]}")

generate_trusted_proxies() {
  {
    for cidr in "${ALL_CIDRS[@]}"; do
      echo "set_real_ip_from ${cidr};"
    done
  } > ci/expected/trusted-proxy-sources.conf
}

generate_origin_allowlist() {
  {
    for cidr in "${ALL_CIDRS[@]}"; do
      echo "${cidr} 1;"
    done
  } > ci/expected/origin-allowlist.conf
}

generate_state() {
  jq -n \
    --arg source_url "http://cfmock/client/v4/ips" \
    --arg etag "${ETAG}" \
    --arg hash "${HASH}" \
    --argjson cidrs "$(printf '%s\n' "${ALL_CIDRS[@]}" | jq -R . | jq -s .)" \
    '{
      source_url: $source_url,
      etag: $etag,
      cidrs: $cidrs,
      hash: $hash
    }' \
    > ci/expected/state.json
}

generate_trusted_proxies
generate_origin_allowlist
generate_state

echo "Generated expected CI artifacts"
