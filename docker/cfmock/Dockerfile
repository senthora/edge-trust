FROM nginx:1.31.0-alpine3.23-slim

RUN apk add --no-cache curl

COPY bin/cfmock /usr/local/bin/cfmock

COPY docker/cfmock/default.conf /etc/nginx/conf.d/default.conf
COPY docker/cfmock/entrypoint.sh /entrypoint.sh

HEALTHCHECK --interval=1s --timeout=2s --retries=3 \
  CMD ["curl", "-f", "http://localhost/health"]

ENTRYPOINT ["/entrypoint.sh"]
