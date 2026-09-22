# Déploiement OjoZone

Ce répertoire provisionne le schéma PostgreSQL/PostGIS et décrit le service Render de test.

## Prérequis

- un projet Neon Free dans une région européenne ;
- PostGIS et `pgcrypto`, activés automatiquement par la migration initiale ;
- le binaire [`golang-migrate`](https://github.com/golang-migrate/migrate), Docker ou Podman ;
- un dépôt Git contenant à terme le `Dockerfile` de l'application.

Le script utilise le binaire `migrate` local en priorité, puis l'image `docker.io/migrate/migrate:v4.19.0` avec Docker ou Podman.

## Démarrer PostgreSQL localement

À la racine du projet :

```sh
make up
```

Cette commande démarre PostgreSQL 16 avec PostGIS, attend que la base soit disponible et applique toutes les migrations. La base locale est accessible avec :

```text
postgresql://ojozone@localhost:5432/ojozone?sslmode=disable&search_path=public
```

Cette base de développement utilise l'authentification `trust` et son port est lié exclusivement à `127.0.0.1`. Cette configuration ne doit jamais être utilisée dans un environnement partagé ou en production.

Commandes utiles :

```sh
make ps                 # état des services
make logs               # journaux PostgreSQL
make migrate-version    # version du schéma
make migrate-down       # annule la dernière migration
make down               # arrête sans supprimer les données
make reset              # supprime le volume et recrée le schéma
```

Le Makefile choisit Docker Compose s'il est installé, sinon Podman Compose. Le choix peut être forcé avec `make up COMPOSE='podman compose'`.

## Provisionner le schéma Neon

Utiliser l'URL **directe** Neon avec TLS pour les migrations. L'URL poolée reste préférable pour `DATABASE_URL` dans l'application.

```sh
export DATABASE_URL='postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require&search_path=public'
./deploy/provision.sh
```

Le script applique toutes les migrations puis affiche la version courante. Il peut être relancé sans rejouer les migrations déjà appliquées.

Pour vérifier la version sans modifier la base :

```sh
DATABASE_URL='postgresql://...' ./deploy/migration/migrate.sh version
```

Pour annuler exactement la dernière migration en environnement de développement :

```sh
DATABASE_URL='postgresql://...' ./deploy/migration/migrate.sh down 1
```

Une migration descendante est destructive. Ne jamais l'exécuter sur une base contenant des données utiles sans sauvegarde et validation préalable.

## Créer une migration

```sh
./deploy/migration/new.sh add_product_aliases
```

Cette commande crée une paire séquentielle dans `deploy/migration/sql` :

```text
000002_add_product_aliases.up.sql
000002_add_product_aliases.down.sql
```

Chaque évolution du schéma doit posséder un chemin `down` réaliste et être testée avec un cycle `up`, `down 1`, `up` sur une base temporaire.

## Déployer sur Render

Le fichier `render.yaml` décrit un unique Web Service Docker gratuit :

1. pousser le dépôt chez un fournisseur Git pris en charge par Render ;
2. créer un Blueprint Render à partir de `deploy/render.yaml` ;
3. renseigner `DATABASE_URL` avec l'URL **poolée** Neon dans les secrets Render ;
4. vérifier le nom public proposé, idéalement `ojozone.onrender.com` ;
5. provisionner le schéma avec `./deploy/provision.sh` avant le premier démarrage applicatif ;
6. vérifier `/health` après le déploiement.

Le Blueprint suppose que le futur `Dockerfile` applicatif se trouve à la racine du dépôt. Tant que ce fichier et l'application n'existent pas, Render ne peut pas terminer le build.

## Variables

Voir `.env.example`. Les valeurs réelles ne doivent jamais être ajoutées au dépôt ni passées directement en argument de ligne de commande lorsque l'historique du shell est conservé.

| Variable | Utilisation |
|---|---|
| `DATABASE_URL` | URL PostgreSQL directe pour les scripts ; URL poolée dans l'application Render |
| `MIGRATE_IMAGE` | Image `golang-migrate` facultative, par défaut `docker.io/migrate/migrate:v4.19.0` |
| `APP_ENV` | Environnement applicatif |
| `PUBLIC_BASE_URL` | URL publique du site |

## Contenu

```text
deploy/
|-- .env.example
|-- README.md
|-- provision.sh
|-- render.yaml
`-- migration/
    |-- migrate.sh
    |-- new.sh
    `-- sql/
        |-- 000001_initial_schema.up.sql
        `-- 000001_initial_schema.down.sql
```
