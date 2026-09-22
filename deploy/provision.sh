#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

if [ -z "${DATABASE_URL:-}" ]; then
	echo "Erreur : fournissez l'URL directe Neon dans DATABASE_URL." >&2
	exit 1
fi

"$SCRIPT_DIR/migration/migrate.sh" up
printf 'Version du schéma : '
"$SCRIPT_DIR/migration/migrate.sh" version
