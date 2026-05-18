import 'go.just'

# Build compose service images
compose-build *args: build
    just compose build {{args}}

# Start local compose stack
compose-up *args:
    just compose build
    just compose up -d {{args}}

# Stop stack and remove volumes/orphans
compose-down:
    just compose down -v --remove-orphans

# Restart compose services
compose-restart:
    just compose restart

# Print compose service logs
compose-logs *args:
    @-just compose logs {{args}}

[private]
run-edge-trust *args:
    @just compose run --rm edge-trust {{args}}

[private]
exec-cfmock *args:
    @just compose exec cfmock cfmock {{args}}

[private]
exec-inspect *args:
    docker exec -it edge-trust-inspect {{args}}

[private]
compose *args:
    COMPOSE_FILE=docker-compose.yml:docker-compose.dev.yml \
    docker compose {{args}}
