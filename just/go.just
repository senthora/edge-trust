# Delete compiled binaries
clean:
    rm -rf bin/*

# Build project binaries
build:
    mkdir -p bin
    just build-target bin/edge-trust ./cmd/edge-trust
    just build-target bin/cfmock ./cmd/cfmock

# Run Go tests
test *args:
    go test {{args}} ./...

[private]
build-target output target:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o {{output}} {{target}}
