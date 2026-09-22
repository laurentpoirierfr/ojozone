#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
MIGRATIONS_DIR="$SCRIPT_DIR/sql"
MIGRATE_IMAGE=${MIGRATE_IMAGE:-docker.io/migrate/migrate:v4.19.0}

if [ -z "${DATABASE_URL:-}" ]; then
	echo "Erreur : DATABASE_URL doit contenir l'URL PostgreSQL cible." >&2
	exit 1
fi

command_name=${1:-up}
shift 2>/dev/null || true

case "$command_name" in
	up|down|goto|version|force)
		;;
	*)
		echo "Usage : DATABASE_URL=... $0 {up|down|goto|version|force} [argument]" >&2
		exit 2
		;;
esac

if command -v migrate >/dev/null 2>&1; then
	exec migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" "$command_name" "$@"
fi

if command -v docker >/dev/null 2>&1; then
	exec docker run --rm \
		-v "$MIGRATIONS_DIR:/migrations:ro" \
		"$MIGRATE_IMAGE" \
		-path=/migrations -database "$DATABASE_URL" "$command_name" "$@"
fi

if command -v podman >/dev/null 2>&1; then
	exec podman run --rm \
		-v "$MIGRATIONS_DIR:/migrations:ro" \
		"$MIGRATE_IMAGE" \
		-path=/migrations -database "$DATABASE_URL" "$command_name" "$@"
fi

echo "Erreur : installez golang-migrate, Docker ou Podman." >&2
exit 1
