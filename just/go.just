# Delete compiled binaries
clean:
    rm -rf bin/*

[arg("target", pattern="all|edge-trust|cfmock")]
build target="all":
    @mkdir -p bin
    @if [ "{{target}}" = "all" ]; then \
        just build-target bin/edge-trust ./cmd/edge-trust; \
        just build-target bin/cfmock ./cmd/cfmock; \
    elif [ "{{target}}" = "edge-trust" ]; then \
        just build-target bin/edge-trust ./cmd/edge-trust; \
    elif [ "{{target}}" = "cfmock" ]; then \
        just build-target bin/cfmock ./cmd/cfmock; \
    fi

# Run Go tests
test *args:
    go test -count=1 {{args}} ./...

[private]
build-target output target:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o {{output}} {{target}}
