#!/usr/bin/env bash
# Lance l'API OjoZone contre une base PostgreSQL de test dediee, applique les
# migrations, installe le referentiel et execute les tests d'integration.
# Usage : scripts/run.sh [flags go test...]
#   ex.  scripts/run.sh -run TestRegister -v
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

DB_NAME="${OJZONE_TEST_DB:-ojozone_test}"
API_PORT="${OJZONE_TEST_PORT:-18081}"
BASE_URL="http://127.0.0.1:${API_PORT}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
ADMIN_EMAIL="${OJZONE_TEST_ADMIN_EMAIL:-admin@example.com}"
ADMIN_PASSWORD="${OJZONE_TEST_ADMIN_PASSWORD:-admin-password-123}"
JWT_SECRET="${OJZONE_TEST_JWT_SECRET:-ojozone-integration-secret-0123456789abcdef}"
# Hash bcrypt de ADMIN_PASSWORD (génére une fois, voir tests/README.md).
ADMIN_PASSWORD_HASH='$2a$10$T9dMQMYiH7RVE1pVrfCw5ergGIntb8DUy4z/IXZVUXxhX/V6sZgdO'
# Hash bcrypt du mot de passe par defaut du moderateur (voir tests/README.md).
MODERATOR_EMAIL="${OJZONE_TEST_MODERATOR_EMAIL:-moderator@example.com}"
MODERATOR_PASSWORD_HASH='$2a$10$NSv/Jl79kvUnTgJ11zyB8ekwV6d/BH6G9DpeS5TwVZ1BvSP4ezXb.'
DATABASE_URL="postgres://ojozone@localhost:${POSTGRES_PORT}/${DB_NAME}?sslmode=disable&search_path=public"
MIGRATE_DATABASE_URL="postgres://ojozone@postgres:5432/${DB_NAME}?sslmode=disable&search_path=public"
SERVER_LOG="${OJZONE_TEST_LOG:-/tmp/ojozone-test-api.log}"
SERVER_BIN="$(mktemp -t ojozone-api.XXXXXX)"

# Choix du moteur de conteneurs, aligne sur la racine du projet.
if command -v docker >/dev/null 2>&1; then
	COMPOSE="docker compose"
elif command -v podman >/dev/null 2>&1; then
	COMPOSE="podman compose"
else
	echo "docker ou podman est requis pour les tests d'integration." >&2
	exit 1
fi

server_pid=""
cleanup() {
	status=$?
	if [ -n "$server_pid" ] && kill -0 "$server_pid" 2>/dev/null; then
		kill "$server_pid" 2>/dev/null || true
		wait "$server_pid" 2>/dev/null || true
	fi
	rm -f "$SERVER_BIN"
	if [ "$status" -ne 0 ] && [ -f "$SERVER_LOG" ]; then
		echo "===== journal du serveur ($SERVER_LOG) =====" >&2
		tail -n 50 "$SERVER_LOG" >&2 || true
	fi
	return "$status"
}
trap cleanup EXIT

echo "==> Demarrage de PostgreSQL"
$COMPOSE up -d postgres
for _ in $(seq 1 30); do
	if $COMPOSE exec -T postgres pg_isready -h 127.0.0.1 -U ojozone >/dev/null 2>&1; then
		break
	fi
	sleep 1
done

echo "==> Recreation de la base $DB_NAME (test isolé)"
$COMPOSE exec -T postgres psql -U ojozone -d postgres -c "DROP DATABASE IF EXISTS \"${DB_NAME}\"" >/dev/null
$COMPOSE exec -T postgres psql -U ojozone -d postgres -c "CREATE DATABASE \"${DB_NAME}\"" >/dev/null

echo "==> Application des migrations sur $DB_NAME"
$COMPOSE run --rm migrate -path=/migrations -database "$MIGRATE_DATABASE_URL" up >/dev/null

echo "==> Installation du referentiel et du compte administrateur"
$COMPOSE exec -T postgres psql -U ojozone -d "$DB_NAME" -v ON_ERROR_STOP=1 >/dev/null <<SQL
INSERT INTO geo_areas (id, type, code, name, country_code) VALUES
    ('2c65c9b4-4527-4e2a-9d2a-6b0a7a1f8c11', 'country', 'FR', 'France', 'FR'),
    ('8d7a3f2d-6a1e-4c2b-9e5b-1f8c0d2a4b6e', 'city', '75056', 'Paris', 'FR')
ON CONFLICT DO NOTHING;

INSERT INTO sources (id, name, kind, is_active) VALUES
    ('0cbdc6bf-361b-4878-b407-e77f735098af', 'Source d''integration', 'official', true),
    ('a1b2c3d4-5e6f-4a8b-9c0d-1e2f3a4b5c6d', 'Contribution citoyenne', 'community', true)
ON CONFLICT DO NOTHING;

INSERT INTO fuel_types (id, code, name_i18n, energy) VALUES
    ('b25fbc0e-9d7a-4d2e-bf5f-7d2a1f0a3c92', 'SP95-E10',
     '{"fr":"Sans plomb 95 - E10","en":"Unleaded 95 - E10"}', 'petrol')
ON CONFLICT DO NOTHING;

INSERT INTO categories (id, slug, name_i18n) VALUES
    ('765c4c3e-9a2e-4a7f-a272-586a311cbb80', 'food', '{"fr":"Alimentation","en":"Food"}')
ON CONFLICT DO NOTHING;

INSERT INTO units (code, dimension, to_base_factor) VALUES
    ('L', 'volume', 1),
    ('kg', 'mass', 1),
    ('u', 'count', 1)
ON CONFLICT DO NOTHING;

INSERT INTO locations (id, name, geo_area_id, address) VALUES
    ('9a80c20f-9055-442d-97d1-43ee336230f0', 'Supermarche integration',
     '8d7a3f2d-6a1e-4c2b-9e5b-1f8c0d2a4b6e', '1 rue du test')
ON CONFLICT DO NOTHING;

INSERT INTO users (email, password_hash, role) VALUES
    ('${ADMIN_EMAIL}', '${ADMIN_PASSWORD_HASH}', 'admin'),
    ('${MODERATOR_EMAIL}', '${MODERATOR_PASSWORD_HASH}', 'moderator')
ON CONFLICT (email) DO UPDATE SET role = EXCLUDED.role;
SQL

echo "==> Compilation du serveur"
(cd "$ROOT/app/backend" && go build -o "$SERVER_BIN" ./cmd/server)

echo "==> Demarrage de l'API sur $BASE_URL"
: > "$SERVER_LOG"
DATABASE_URL="$DATABASE_URL" PORT="$API_PORT" AUTH_JWT_SECRET="$JWT_SECRET" \
	AUTH_ACCESS_TTL=1h AUTH_ISSUER=ojozone-test "$SERVER_BIN" >>"$SERVER_LOG" 2>&1 &
server_pid=$!

ready=0
for _ in $(seq 1 40); do
	if curl -fsS "$BASE_URL/ops/liveness" >/dev/null 2>&1; then
		ready=1
		break
	fi
	if ! kill -0 "$server_pid" 2>/dev/null; then
		echo "Le serveur s'est arrete avant d'etre pret." >&2
		exit 1
	fi
	sleep 0.25
done
if [ "$ready" != "1" ]; then
	echo "L'API n'est pas prete sur $BASE_URL." >&2
	exit 1
fi

echo "==> Execution des tests d'integration"
(cd "$ROOT/tests" && \
OJZONE_TEST_BASE_URL="$BASE_URL" \
OJZONE_TEST_ADMIN_EMAIL="$ADMIN_EMAIL" \
OJZONE_TEST_ADMIN_PASSWORD="$ADMIN_PASSWORD" \
go test -count=1 "$@" ./integration/...)