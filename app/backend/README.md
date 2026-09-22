# Backend OjoZone

API Go basée sur Gin, PostgreSQL/pgx et le code d'accès aux données généré par sqlc.dev.

## Démarrage local

Depuis `app/backend` :

```sh
make db-up
make run
```

Le serveur écoute par défaut sur `http://localhost:8080`.

## Documentation API

| Ressource | URL |
|---|---|
| Swagger UI | `http://localhost:8080/swagger/index.html` |
| Spécification JSON | `http://localhost:8080/swagger/doc.json` |
| Spécification versionnée | `docs/swagger.json` et `docs/swagger.yaml` |

La documentation est générée avec [Swaggo](https://github.com/swaggo/swag) à partir des annotations Go :

```sh
make generate-swagger
```

`make generate` régénère à la fois le code sqlc et Swagger. `make check` régénère les artefacts, formate, analyse et teste le backend.

Toute nouvelle route doit comporter au minimum `@Summary`, `@Tags`, `@Produce`, ses paramètres, ses réponses et `@Router`. Les fichiers du dossier `docs/` doivent être régénérés et ajoutés avec la modification.

## Routes

```text
GET /ops/liveness
GET /ops/readiness
GET /ops/infos

POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout

GET /api/v1/me
PATCH /api/v1/me
DELETE /api/v1/me
GET /api/v1/me/contributions

GET /api/v1/products
POST /api/v1/products
GET /api/v1/products/{id}
PUT /api/v1/products/{id}
DELETE /api/v1/products/{id}
GET /api/v1/products/by-barcode/{barcode}
GET /api/v1/products/{id}/prices

POST /api/v1/product-prices
GET /api/v1/product-prices
GET /api/v1/product-prices/{id}
PUT /api/v1/product-prices/{id}
DELETE /api/v1/product-prices/{id}
```

`POST /api/v1/products` est idempotent au niveau métier : il effectue un upsert par code-barres lorsque celui-ci est fourni, sinon par le champ `id`. `PUT` effectue l'upsert avec l'UUID présent dans le chemin. `DELETE` refuse la suppression avec `409 Conflict` lorsque le produit est encore référencé.

`POST /api/v1/product-prices` effectue un upsert par `(source_id, source_record_id)`. Les montants et quantités sont transmis sous forme de chaînes décimales pour éviter toute perte de précision. `PUT /api/v1/product-prices/{id}` effectue l'upsert par UUID.

Les erreurs sont retournées avec le type `application/problem+json`.

## Authentification et autorisation

L'API protège ses écritures par jeton d'accès JWT (en-tête `Authorization: Bearer <token>`) émis
avec un jeton de rafraîchissement opaque tournant à chaque usage.

- Rôles : `member` (par défaut), `moderator`, `admin`, `partner`.
- Écritures du catalogue et des référentiels : rôle `admin`.
- Ajout d'événements de modération : rôles `moderator` et `admin`.
- Administration des utilisateurs (`GET /api/v1/admin/*`) : rôle `admin`.
- Lecture publique, inscription, connexion, renouvellement et profil (`/api/v1/me*`) : aucun rôle requis.

Variables d'environnement :

| Variable | Défaut | Description |
|---|---|---|
| `AUTH_JWT_SECRET` | clé éphémère (dev) | Clé HMAC-SHA256 des jetons d'accès ; obligatoire en production |
| `AUTH_ISSUER` | `ojozone-api` | Émetteur porté par les jetons |
| `AUTH_ACCESS_TTL` | `15m` | Durée de vie d'un jeton d'accès |

Les jetons de rafraîchissement expirent après 30 jours ; les sessions révoquées ou réutilisées
sont invalidées. `DELETE /api/v1/me` anonymise l'utilisateur et révoque toutes ses sessions.
