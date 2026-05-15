set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

import 'just/go.just'
import 'just/docker.just'

# List all available recipes
default:
    @just --list

# Inspect generated artifacts and state
inspect:
    @echo '=== Trusted Proxy Sources ==='
    @just show-trusted-proxies

    @echo
    @echo '=== Origin Allowlist ==='
    @just show-origin-allowlist

    @echo
    @echo '=== State ==='
    @just show-state

# Randomize mock Cloudflare CIDRs
randomize-cidrs:
    @just exec-cfmock-cmd cfmock random

# Clear all mock Cloudflare CIDRs
clear-cidrs:
    @just exec-cfmock-cmd cfmock clear

# Delete mock Cloudflare API response
delete-response:
    @just exec-cfmock-cmd cfmock delete

# Set mock Cloudflare IPv4 CIDRs
set-ipv4 *args:
    @just exec-cfmock-cmd IPV4_CIDRS='{{args}}' cfmock set

# Set mock Cloudflare IPv6 CIDRs
set-ipv6 *args:
    @just exec-cfmock-cmd IPV6_CIDRS='{{args}}' cfmock set

# Set mock Cloudflare ETag
set-etag *args:
    @just exec-cfmock-cmd ETAG='{{args}}' cfmock set

# Print generated trusted proxy config
show-trusted-proxies:
    @just exec-edge-trust-cmd \
    'cat /etc/nginx/dynamic/trusted-proxy-sources.conf'

# Print generated origin allowlist config
show-origin-allowlist:
    @just exec-edge-trust-cmd \
    'cat /etc/nginx/dynamic/origin-allowlist.conf'

# Print persisted reconciliation state
show-state:
    @just exec-edge-trust-cmd \
    'cat /var/lib/edge-trust/state.json'
