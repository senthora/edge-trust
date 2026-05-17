# CFMock

CFMock is a lightweight Cloudflare IP API emulator for local development and integration testing.
It serves Cloudflare-compatible IP range responses over HTTP using nginx and a small Go control utility.

It is primarily used for testing EdgeTrust update and configuration workflows,
but can also be used standalone anywhere a mock Cloudflare IP API is needed.

## Overview

CFMock exposes a minimal HTTP API compatible with Cloudflare IP ranges endpoint:
> https://api.cloudflare.com/client/v4/ips

It consists of two components:

- Nginx serving static JSON responses
- Go utility for mutating response state

The mock response is stored as a JSON file on disk and served directly by nginx.

Updates made through `cfmock` CLI are immediately reflected over HTTP without requiring reloads or restarts.

This makes it useful for testing:

- update detection
- ETag changes
- nginx configuration regeneration
- empty upstream states
- upstream failure scenarios

## Quick Start

Run CFMock using Docker Compose:

```yaml
services:
  cfmock:
    image: senthora/cfmock:latest
    ports:
      - "80:80"
    tmpfs:
      - /usr/share/nginx/html
    environment:
      ETAG: a8e453d9d129a3769407127936edfdb0
      IPV4_CIDRS: 199.27.128.0/21
      IPV6_CIDRS: 2400:cb00::/32
    restart: unless-stopped
```

Start container:

```bash
docker compose up -d
```

Verify response:

```bash
curl http://localhost/client/v4/ips
```

## Usage

### Commands

| Command    | Description                          |
|------------|--------------------------------------|
| `set`      | Set exact response values            |
| `random`   | Generate random ETag and IP ranges   |
| `clear`    | Clear all IP ranges and ETag         |
| `delete`   | Delete the mock response file        |

### Flags

| Flag       | Description                |
|------------|----------------------------|
| `--etag`   | Exact ETag value           |
| `--ipv4`   | IPv4 CIDR (repeatable)     |
| `--ipv6`   | IPv6 CIDR (repeatable)     |

### Environment Variables

| Variable       | Description                  |
|----------------|------------------------------|
| `ETAG`         | Response ETag                |
| `IPV4_CIDRS`   | Comma-separated IPv4 CIDRs   |
| `IPV6_CIDRS`   | Comma-separated IPv6 CIDRs   |

## Examples

Set exact response values:

```shell
cfmock set \
  --etag abc123 \
  --ipv4 173.245.48.0/20 \
  --ipv4 103.21.244.0/22 \
  --ipv6 2400:cb00::/32
```

> [!NOTE]
> IP flags are repeatable and can be used multiple times.

Configure response values using environment variables:

```shell
ETAG=abc123 \
IPV4_CIDRS=173.245.48.0/20,103.21.244.0/22 \
IPV6_CIDRS=2400:cb00::/32 \
cfmock set
```

> [!NOTE]
> Flags override environment variables when both are provided.

Generate random ETag and IP ranges:

```shell
cfmock random
```

Clear all IP ranges and ETag:

```shell
cfmock clear
```

Delete the mock response file:

```shell
cfmock delete
```

> [!IMPORTANT] When running in a container, `cfmock set` is executed automatically on startup.
> This allows response state to be initialized directly from environment variables.

## HTTP API

| Endpoint         | Description               |
|------------------|---------------------------|
| `/client/v4/ips` | Mock Cloudflare IP API    |
| `/health`        | Container health endpoint |

All other endpoints return `404`.

**Example response:**

```json
{
  "errors": null,
  "messages": null,
  "success": false,
  "result": {
    "etag": "a8e453d9d129a3769407127936edfdb0",
    "ipv4_cidrs": [
      "199.27.128.0/21"
    ],
    "ipv6_cidrs": [
      "2400:cb00::/32"
    ]
  }
}
```

## Development

### Requirements

- Go
- Docker
- Docker Compose
- Just

### Setup

Build CFMock binary:

```bash
just build cfmock
```

Start local development stack:

```bash
just compose-up
```

Stop development stack:

```bash
just compose-down
```

For more information run `just` to list all available commands.
