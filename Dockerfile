FROM alpine:3.23

COPY bin/edge-trust /usr/local/bin/edge-trust

HEALTHCHECK --start-period=2s --interval=10s --timeout=3s --retries=3 \
  CMD ["edge-trust", "healthcheck"]
