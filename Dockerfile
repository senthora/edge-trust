FROM gcr.io/distroless/static-debian13:latest

COPY build/bin/edge-trust /usr/local/bin/edge-trust

ENTRYPOINT ["/usr/local/bin/edge-trust"]
