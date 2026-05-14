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
