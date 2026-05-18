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

# Run Go tests with coverage
test-coverage:
    @mkdir -p build/coverage
    go test \
      -count=1 \
      -coverprofile=build/coverage/coverage.out \
      `go list ./... | grep -Ev '/cmd/cfmock|/testutil'`

    go tool cover -html=build/coverage/coverage.out \
      -o build/coverage/coverage.html

# Print test coverage report
print-coverage:
    go tool cover -func=build/coverage/coverage.out

[private]
build-target output target:
    @mkdir -p build/bin
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o {{output}} {{target}}
