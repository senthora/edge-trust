# Edge Trust

Edge Trust helps protect infrastructure running
behind Cloudflare by managing nginx trust configuration.

## Why do I need this?

When running infrastructure behind Cloudflare, origin servers 
are expected to only accept traffic forwarded by Cloudflare proxies.

If origin servers remain publicly reachable without additional restrictions,
attackers can bypass Cloudflare entirely and connect directly to the origin.
This can expose services to attacks, origin IP discovery,
and traffic that bypasses Cloudflare security controls.

Edge Trust helps maintain a "dark origin" setup by continuously synchronizing
nginx trust configuration with the latest published Cloudflare IP ranges.

This allows nginx to:

- trust real client IP headers only from Cloudflare proxies
- restrict direct origin access to Cloudflare networks
- automatically stay up to date as Cloudflare IP ranges change

## How does it work?

Edge Trust periodically fetches Cloudflare IP ranges and regenerates nginx
trust configuration when changes are detected. Generated configuration files 
are then written to a shared volume and a reload signal file is created for nginx.

Your nginx container is expected to:

- mount the generated configuration volume
- include generated configuration files
- watch for reload signal files
- reload nginx when signals appear

Edge Trust does not reload nginx directly.

## Quick Start

Define Edge Trust as a docker compose service:

```yaml
edge-trust:
  image: senthora/edge-trust:latest
  volumes:
    - nginx-dynamic:/etc/nginx/dynamic
    - trust-state:/var/lib/edge-trust
    - nginx-signal:/signal
  environment:
    CF_API_URL: https://api.cloudflare.com/client/v4/ips
    STATE_JSON_PATH: /var/lib/edge-trust/state.json
    NGINX_PROXY_SOURCES_PATH: /etc/nginx/dynamic/trusted-proxy-sources.conf
    NGINX_ORIGIN_ALLOWLIST_PATH: /etc/nginx/dynamic/origin-allowlist.conf
    NGINX_RELOAD_SIGNAL_PATH: /signal/nginx.reload

volumes:
  nginx-dynamic:
  trust-state:
  nginx-signal:
```

Then run the binary inside the container:

```shell
docker compose run --rm edge-trust run
```

Generated nginx configuration files will be written to:

```text
/etc/nginx/dynamic
```

Edge Trust can also be run in daemon mode:

```yaml
services:
  edge-trust:
    image: senthora/edge-trust:latest
    command:
      - --daemon
      - --interval=4h
      - run
    healthcheck:
      test: ["CMD", "edge-trust", "healthcheck"]
      interval: 30s
      timeout: 3s
      retries: 3
    restart: unless-stopped
```

This keeps Edge Trust running continuously 
and checks for Cloudflare IP range updates every 4 hours.

Read more about how to setup CFMock [here](https://github.com/senthora/edge-trust/blob/master/cmd/cfmock/README.md).

## Usage

### Commands

| Command       | Description               |
|---------------|---------------------------|
| `run`         | Run update reconciliation |
| `healthcheck` | Validate daemon heartbeat |

### Flags

| Flag            | Description                        | Default |
|-----------------|------------------------------------|---------|
| `--daemon`      | Run continuously instead of once   | `false` |
| `--interval`    | Interval between update checks     | `12h`   |
| `--hb-interval` | Interval between heartbeat updates | `1s`    |
| `--debug`       | Enable debug logging               | `false` |

### Environment Variables

| Variable                      | Description                       |
|-------------------------------|-----------------------------------|
| `CF_API_URL`                  | Cloudflare IP ranges endpoint     |
| `NGINX_PROXY_SOURCES_PATH`    | Trusted proxy sources output path |
| `NGINX_ORIGIN_ALLOWLIST_PATH` | Origin allowlist output path      |
| `NGINX_RELOAD_SIGNAL_PATH`    | Nginx reload signal file path     |
| `STATE_JSON_PATH`             | State file path                   |
| `HEALTH_SIGNAL_PATH`          | Heartbeat signal file path        |

## Development

### Requirements

- [Go](https://go.dev/)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Just](https://github.com/casey/just)

### Quick Start

Build and start development stack:

```shell
just compose-up
```

Inspect generated configuration and state:

```shell
just inspect
```

Stop development stack:

```shell
just compose-down
```

### Testing

Run unit tests:

```shell
just test
```

Manually modify response and inspect result:

```shell
just compose-up
just randomize-cidrs
just set-etag <sample-etag>
just inspect
```

Run integration tests locally:

```shell
export IPV4_CIDRS=103.21.244.0/22 \
  IPV6_CIDRS=2400:cb00::/32 \
  ETAG=7f4c2d91e6ab3f0c58d2a4b9f1e7c635 \
  HASH=sha256:4e25067cfc8579af9fd81c407bae411d1c6db84de3dd14f32647a3a560d4ea27

docker compose up -d cfmock --wait
docker compose run --rm --no-deps edge-trust --debug run
docker compose run --rm itest
```

## Nginx Configuration

### Origin Allowlist

Used to restrict origin access to Cloudflare IP ranges.

**Example:**

```
# Generated by edge-trust at: 2026-05-14T12:23:36Z
# Source: https://api.cloudflare.com/client/v4/ips
# Total IPs: 4

186.146.35.0/24 1;
2400:4870::/32 1;
2400:9d60::/32 1;
40.87.245.0/24 1;
```

### Proxy Sources

Used to define trusted proxy sources for client IP extraction.

**Example:**

```
# Generated by edge-trust at: 2026-05-14T12:23:36Z
# Source: https://api.cloudflare.com/client/v4/ips
# Total IPs: 4

set_real_ip_from 186.146.35.0/24;
set_real_ip_from 2400:4870::/32;
set_real_ip_from 2400:9d60::/32;
set_real_ip_from 40.87.245.0/24;
```

## License

This project is licensed under MIT License.
