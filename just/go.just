# Delete compiled binaries
clean:
    @rm -rf build/*

# Build all binaries
build: build-edge-trust build-cfmock

# Build edge-trust binary
build-edge-trust:
    just build-target build/bin/edge-trust ./cmd/edge-trust

# Build cfmock binary
build-cfmock:
    just build-target build/bin/cfmock ./cmd/cfmock

# Run Go tests
test *args:
    go test -count=1 {{args}} ./...

[private]
build-target output target:
    @mkdir -p build/bin
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o {{output}} {{target}}
