SHELL := /bin/sh

COMPOSE ?= $(shell if command -v docker >/dev/null 2>&1; then echo "docker compose"; elif command -v podman >/dev/null 2>&1; then if printf '%s' "$$XDG_DATA_HOME" | grep -q '/snap/code/'; then echo "env XDG_DATA_HOME=$$HOME/.local/share podman compose"; else echo "podman compose"; fi; else echo "docker compose"; fi)
CONTAINER_ENGINE ?= $(shell if command -v docker >/dev/null 2>&1; then echo "docker"; elif command -v podman >/dev/null 2>&1; then if printf '%s' "$$XDG_DATA_HOME" | grep -q '/snap/code/'; then echo "env XDG_DATA_HOME=$$HOME/.local/share podman"; else echo "podman"; fi; else echo "docker"; fi)
DB_URL ?= postgres://ojozone@localhost:5432/ojozone?sslmode=disable&search_path=public
DB_URL_CONTAINER ?= postgres://ojozone@postgres:5432/ojozone?sslmode=disable&search_path=public
GOPLANTUML ?= goplantuml
DOMAIN_MODEL ?= assets/domain-model.puml
DOMAIN_MODEL_PNG ?= assets/domain-model.png
PLANTUML_IMAGE ?= docker.io/plantuml/plantuml:1.2026.6

.PHONY: help up db-up db-ready down restart ps logs migrate migrate-down migrate-version migration-new reset domain-model domain-model-source domain-model-png test-integration

help: ## Afficher les commandes disponibles
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: migrate ## Démarrer PostgreSQL puis appliquer les migrations

db-up:
	$(COMPOSE) up -d postgres

db-ready: db-up
	@$(COMPOSE) exec -T postgres sh -c 'until pg_isready -h 127.0.0.1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" >/dev/null 2>&1; do sleep 1; done'

migrate: db-ready ## Appliquer toutes les migrations en attente
	$(COMPOSE) run --rm migrate -path=/migrations -database '$(DB_URL_CONTAINER)' up

down: ## Arrêter les services sans supprimer les données
	$(COMPOSE) down

restart: down up ## Redémarrer PostgreSQL et appliquer les migrations

ps: ## Afficher l'état des services
	$(COMPOSE) ps

logs: ## Suivre les journaux PostgreSQL
	$(COMPOSE) logs -f postgres

migrate-down: db-ready ## Annuler la dernière migration (développement uniquement)
	$(COMPOSE) run --rm migrate -path=/migrations -database '$(DB_URL_CONTAINER)' down 1

migrate-version: db-ready ## Afficher la version du schéma local
	$(COMPOSE) run --rm migrate -path=/migrations -database '$(DB_URL_CONTAINER)' version

migration-new: ## Créer une migration : make migration-new NAME=add_table
	@test -n "$(NAME)" || (echo "Usage : make migration-new NAME=nom_de_la_migration" >&2; exit 2)
	./deploy/migration/new.sh "$(NAME)"

domain-model: domain-model-png ## Générer le modèle du domaine en PlantUML et PNG

domain-model-source: ## Générer la source PlantUML du package domain
	@command -v "$(GOPLANTUML)" >/dev/null 2>&1 || (echo "goplantuml est introuvable dans PATH" >&2; exit 1)
	@mkdir -p "$(dir $(DOMAIN_MODEL))"
	$(GOPLANTUML) \
		-title "OjoZone domain model" \
		-show-aggregations \
		-show-compositions \
		-show-aliases \
		-show-connection-labels \
		-output "$(DOMAIN_MODEL)" \
		./app/backend/internal/domain

domain-model-png: domain-model-source ## Transformer le modèle PlantUML en PNG
	@set -eu; \
		tmp="$(DOMAIN_MODEL_PNG).tmp"; \
		trap 'rm -f "$$tmp"' EXIT; \
		$(CONTAINER_ENGINE) run --rm -i "$(PLANTUML_IMAGE)" -tpng -pipe < "$(DOMAIN_MODEL)" > "$$tmp"; \
		test -s "$$tmp"; \
		mv "$$tmp" "$(DOMAIN_MODEL_PNG)"; \
		trap - EXIT

reset: ## Supprimer les données locales et recréer le schéma
	$(COMPOSE) down -v
	$(MAKE) up

test-integration: ## Lancer les tests d'intégration API (docker/podman requis)
	$(MAKE) -C tests check
