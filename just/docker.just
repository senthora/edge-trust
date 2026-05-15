import 'go.just'

# Build compose service images
compose-build *args: build
    docker compose build {{args}}

# Start local compose stack
compose-up:
    just compose-build
    docker compose up -d

# Stop stack and remove volumes/orphans
compose-down:
    docker compose down -v --remove-orphans

# Restart compose services
compose-restart:
    docker compose restart

# Print compose service logs
compose-logs *args:
    @-docker compose logs {{args}}

[private]
exec-edge-trust-it:
    @docker compose exec -it edge-trust sh

[private]
exec-edge-trust-cmd +args:
    @docker compose exec edge-trust sh -ceu '{{args}}'

[private]
exec-cfmock-it:
    @docker compose exec -it cfmock sh

[private]
exec-cfmock-cmd +args:
    @docker compose exec cfmock sh -ceu '{{args}}'
