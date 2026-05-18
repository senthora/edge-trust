# CFMock

CFMock is a lightweight Cloudflare IP API emulator for local development and integration testing.

It serves Cloudflare-compatible IP range responses over HTTP using nginx and a small Go control utility.

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

```shell
docker compose up -d
```

Verify response:

```shell
curl http://localhost/client/v4/ips
```

## Endpoints

| Endpoint         | Description               |
|------------------|---------------------------|
| `/client/v4/ips` | Mock Cloudflare IP API    |
| `/health`        | Container health endpoint |

All other endpoints return `404`.

## Environment Variables

| Variable       | Description                  |
|----------------|------------------------------|
| `ETAG`         | Response ETag                |
| `IPV4_CIDRS`   | Comma-separated IPv4 CIDRs   |
| `IPV6_CIDRS`   | Comma-separated IPv6 CIDRs   |

> [!IMPORTANT]
> When running in a container, `cfmock set` is executed automatically on startup.
> This allows response state to be initialized directly from environment variables.

## Example Response

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

For full documentation read the CFMock
[README](https://github.com/senthora/edge-trust/blob/master/cmd/cfmock/README.md).
