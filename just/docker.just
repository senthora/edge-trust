import 'go.just'

# Build compose service images
compose-build *args: build
    just compose build {{args}}

# Start local compose stack
compose-up:
    just compose build
    just compose up -d

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
exec-edge-trust-it:
    @just compose exec -it edge-trust sh

[private]
exec-edge-trust-cmd +args:
    @just compose exec edge-trust sh -ceu '{{args}}'

[private]
exec-cfmock-it:
    @just compose exec -it cfmock sh

[private]
exec-cfmock-cmd +args:
    @just compose exec cfmock sh -ceu '{{args}}'

[private]
compose *args:
    COMPOSE_FILE=docker-compose.yml:docker-compose.dev.yml \
    docker compose {{args}}
