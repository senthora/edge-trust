FROM nginx:1.30.1-alpine3.23

RUN apk add --no-cache curl

COPY bin/cfmock /usr/local/bin/cfmock

COPY docker/cfmock/ips.json /usr/share/nginx/html/ips.json
COPY docker/cfmock/default.conf /etc/nginx/conf.d/default.conf
COPY docker/cfmock/entrypoint.sh /entrypoint.sh

HEALTHCHECK --interval=1s --timeout=2s --retries=3 \
  CMD ["curl", "-f", "http://localhost/health"]

ENTRYPOINT ["/entrypoint.sh"]
