# Edge Trust

Edge Trust keeps nginx trust configuration in sync with Cloudflare
IP ranges to help protect dark origin infrastructure behind Cloudflare.

## Overview

It periodically fetches Cloudflare IP ranges and keeps 
nginx trust configuration synchronized with the latest published CIDRs.

On each update cycle it:

- fetches Cloudflare IPv4 and IPv6 ranges
- validates and normalizes CIDRs
- compares the fetched ETag against persisted state
- skips regeneration if nothing changed
- generates nginx trust configuration files
- atomically writes updated configuration artifacts
- persists canonical synchronization state
- emits a filesystem reload signal for nginx

If an update fails, the last known good state and generated configuration
are preserved until the next successful reconciliation cycle.

