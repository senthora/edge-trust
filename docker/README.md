# Edge Trust

When running infrastructure behind Cloudflare, origin servers 
are expected to only accept traffic forwarded by Cloudflare proxies.

Without additional restrictions, attackers can 
bypass Cloudflare entirely and connect directly to the origin.

Edge Trust helps maintain a "dark origin" setup by keeping nginx 
trust configuration up to date with the latest published Cloudflare IP ranges.

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

Run Edge Trust using Docker Compose:

```yaml
services:
  edge-trust:
    image: senthora/edge-trust:latest
    command:
      - --daemon
      - --interval=4h
      - run
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
    healthcheck:
      test: ["CMD", "edge-trust", "healthcheck"]
      interval: 30s
      timeout: 3s
      retries: 3
    restart: unless-stopped

volumes:
  nginx-dynamic:
  trust-state:
  nginx-signal:
```

Start container:

```shell
docker compose up -d
```

Generated nginx configuration files will be written to:

```text
/etc/nginx/dynamic
```

For full documentation read Edge Trust [README](https://github.com/senthora/edge-trust/blob/master/README.md).
