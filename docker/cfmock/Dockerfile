FROM nginx:1.31.0-alpine3.23-slim

COPY bin/cfmock /usr/local/bin/cfmock

COPY docker/cfmock/default.conf /etc/nginx/conf.d/default.conf
COPY docker/cfmock/entrypoint.sh /entrypoint.sh

HEALTHCHECK --interval=1s --timeout=2s --retries=3 \
  CMD ["wget", "--spider", "-q", "http://127.0.0.1/health"]

ENTRYPOINT ["/entrypoint.sh"]
