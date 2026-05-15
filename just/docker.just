import 'go.just'

# Build compose service images
compose-build *args: build
    docker compose build {{args}}

# Start local compose stack
compose-up:
    just compose-build
    docker compose up -d

# Stop stack and remove volumes/orphans
compose-down:
    docker compose down -v --remove-orphans

# Restart compose services
compose-restart:
    docker compose restart

# Print compose service logs
compose-logs *args:
    @-docker compose logs {{args}}

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
    cat /etc/nginx/dynamic/trusted-proxy-sources.conf

# Print generated origin allowlist config
show-origin-allowlist:
    @just exec-edge-trust-cmd \
    cat /etc/nginx/dynamic/origin-allowlist.conf

# Print persisted reconciliation state
show-state:
    @just exec-edge-trust-cmd \
    cat /var/lib/edge-trust/state.json

[private]
exec-edge-trust-it:
    @docker compose exec -it edge-trust sh

[private]
exec-edge-trust-cmd +args:
    @docker compose exec edge-trust sh -ceu '{{args}}'

[private]
exec-cfmock-it:
    @docker compose exec -it cfmock sh

[private]
exec-cfmock-cmd +args:
    @docker compose exec cfmock sh -ceu '{{args}}'
