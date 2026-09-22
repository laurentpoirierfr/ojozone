#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
	echo "Usage : $0 nom_de_la_migration" >&2
	exit 2
fi

case "$1" in
	*[!a-z0-9_]*)
		echo "Erreur : utilisez uniquement des minuscules, chiffres et underscores." >&2
		exit 2
		;;
esac

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

if command -v migrate >/dev/null 2>&1; then
	exec migrate create -ext sql -dir "$SCRIPT_DIR/sql" -seq -digits 6 "$1"
fi

echo "Erreur : le binaire migrate est requis pour créer une migration." >&2
echo "Installation : https://github.com/golang-migrate/migrate/tree/master/cmd/migrate" >&2
exit 1
