SHELL := /bin/sh
COMPOSE ?= docker compose

.PHONY: help check test down logs

help: ## Afficher les commandes disponibles
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "%-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check: ## Lancer PostgreSQL de test, l'API et la suite d'integration
	./scripts/run.sh

test: check ## Alias de check

down: ## Arreter PostgreSQL (les données de test sont conservees)
	$(COMPOSE) down

logs: ## Suivre le journal du serveur de test
	tail -f /tmp/ojozone-test-api.log