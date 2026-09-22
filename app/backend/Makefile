SHELL := /bin/sh

GO ?= go
SQLC ?= sqlc
SWAG ?= swag
BINARY ?= bin/ojozone-api
DATABASE_URL ?= postgres://ojozone@localhost:5432/ojozone?sslmode=disable&search_path=public

.PHONY: help generate generate-sqlc generate-swagger fmt vet test check build run clean db-up db-down migrate migrate-down migrate-version migration-new

help: ## Afficher les commandes disponibles
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

generate: generate-sqlc generate-swagger ## Générer le code sqlc et la documentation Swagger

generate-sqlc: ## Générer le code Go depuis les requêtes sqlc
	$(SQLC) generate

generate-swagger: ## Générer docs.go, swagger.json et swagger.yaml
	$(SWAG) init -g cmd/server/main.go -o docs --parseInternal

fmt: ## Formater le code Go
	$(GO) fmt ./...

vet: ## Exécuter l'analyse statique Go
	$(GO) vet ./...

test: ## Exécuter les tests Go
	$(GO) test ./...

check: generate fmt vet test ## Générer, formater et valider le backend

build: generate ## Compiler le serveur dans bin/
	@mkdir -p $(dir $(BINARY))
	$(GO) build -o $(BINARY) ./cmd/server

run: generate ## Lancer le serveur avec la base locale
	DATABASE_URL='$(DATABASE_URL)' $(GO) run ./cmd/server

clean: ## Supprimer les artefacts de compilation
	rm -rf bin

db-up: ## Démarrer et migrer PostgreSQL depuis la racine du projet
	$(MAKE) -C ../.. up

db-down: ## Arrêter PostgreSQL sans supprimer les données
	$(MAKE) -C ../.. down

migrate: ## Appliquer les migrations locales
	$(MAKE) -C ../.. migrate

migrate-down: ## Annuler la dernière migration locale
	$(MAKE) -C ../.. migrate-down

migrate-version: ## Afficher la version du schéma local
	$(MAKE) -C ../.. migrate-version

migration-new: ## Créer une migration : make migration-new NAME=add_table
	@test -n "$(NAME)" || (echo "Usage : make migration-new NAME=nom_de_la_migration" >&2; exit 2)
	$(MAKE) -C ../.. migration-new NAME='$(NAME)'
