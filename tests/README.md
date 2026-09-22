# Tests d'intégration

La suite `tests/integration` vérifie l'API OjoZone **de bout en bout** contre une
base PostgreSQL/PostGIS réelle : authenticité des parcours HTTP (statuts, corps
JSON, en-têtes), persistance des données et contrôle d'accès par rôle.

Elle vit dans son propre module Go (`tests/go.mod`) et n'importe **aucun code du
backend** : elle parle uniquement le contrat HTTP public, comme le ferait un
client externe.

## Architecture

```text
tests/
├── go.mod                     # module indépendant, sans dépendances externes
├── Makefile                   # check / down / logs
├── scripts/run.sh             # orchestration : base, migrations, seed, serveur, tests
├── internal/api/              # client HTTP typé + DTO du contrat public
│   ├── client.go              # requêtes, enveloppe {"data":...}, Problem
│   └── methods.go             # méthodes métier (Register, Login, Products…)
└── integration/
    ├── conftest_test.go       # TestMain (vivacité) + helpers, compte admin
    ├── auth_test.go           # inscription, connexion, rotation, profil
    ├── products_test.go       # catalogue, upserts, prix, rôles
    └── errors_test.go         # problem+json, 401/403/404, corps invalides
```

## Prérequis

- Go ≥ 1.26
- Docker (ou Podman) avec Docker Compose, comme le reste du projet

## Exécution

Depuis la racine du projet :

```sh
make test-integration
```

Depuis `tests/` :

```sh
make check        # équivalent à scripts/run.sh
```

`scripts/run.sh` fait, dans l'ordre :

1. Démarre PostgreSQL (`docker compose up -d postgres`) et attend sa disponibilité.
2. Crée la base dédiée `ojozone_test` (variable `OJZONE_TEST_DB`) si absente.
3. Applique les migrations `deploy/migration/sql` avec `golang-migrate`.
4. Installe le référentiel de test et le compte administrateur (voir plus bas).
5. Compile et lance le serveur sur le port `18081` (variable `OJZONE_TEST_PORT`).
6. Exécute `go test -count=1 ./integration/...` (les arguments supplémentaires
   sont transmis à `go test`, ex. `scripts/run.sh -run TestRegister -v`).
7. Arrête le serveur ; PostgreSQL reste actif pour un prochain passage.

:information_source: Sans serveur joignable, la suite refuse de démarrer et
affiche la commande à lancer. C'est voulu : un test d'intégration qui « ignore »
faute d'infrastructure masque une régression.

## Données de test

La base `ojozone_test` reçoit des fixtures déterministes, référencées par les
tests avec des UUID fixes :

| Élément | UUID | Utilisé par |
|---|---|---|
| Zone pays France | `2c65c9b4-4527-4e2a-9d2a-6b0a7a1f8c11` | géographie |
| Zone ville Paris | `8d7a3f2d-6a1e-4c2b-9e5b-1f8c0d2a4b6e` | lieux |
| Source officielle | `0cbdc6bf-361b-4878-b407-e77f735098af` | prix |
| Catégorie « food » | `765c4c3e-9a2e-4a7f-a272-586a311cbb80` | produits |
| Lieu supermarché | `9a80c20f-9055-442d-97d1-43ee336230f0` | prix |
| Unités L, kg, u | — | produits et prix |

Le compte administrateur `admin@example.com` / `admin-password-123` (rôle
`admin`) est inséré à chaque passe. Les tests d'écriture (produits, prix,
administration) le connectent via `OJZONE_TEST_ADMIN_EMAIL` /
`OJZONE_TEST_ADMIN_PASSWORD`.

Pour changer le mot de passe administrateur, régénérez son hash bcrypt et
mettez-le dans `scripts/run.sh` :

```sh
cd app/backend && cat <<'EOF' > /tmp/hash.go
package main
import ("fmt"; "golang.org/x/crypto/bcrypt")
func main() { h, _ := bcrypt.GenerateFromPassword([]byte("nouveau-mot-de-passe"), bcrypt.DefaultCost); fmt.Println(string(h)) }
EOF
go run /tmp/hash.go
```

## Variables d'environnement

| Variable | Défaut | Rôle |
|---|---|---|
| `OJZONE_TEST_DB` | `ojozone_test` | Nom de la base de test |
| `OJZONE_TEST_PORT` | `18081` | Port de l'API |
| `OJZONE_TEST_BASE_URL` | `http://127.0.0.1:18081` | URL lue par la suite Go |
| `OJZONE_TEST_ADMIN_EMAIL` / `OJZONE_TEST_ADMIN_PASSWORD` | `admin@example.com` / `admin-password-123` | Compte administrateur |
| `OJZONE_TEST_JWT_SECRET` | fixe | Clé de signature des jetons |
| `OJZONE_TEST_LOG` | `/tmp/ojozone-test-api.log` | Journal du serveur |

## Ajouter un test

1. Si le parcours mobilise un nouvel endpoint, ajoutez une méthode typée dans
   `tests/internal/api/methods.go` (succès → type décodé + `*Problem`, erreur →
   `*Problem` renseigné).
2. Écrivez le test dans `integration/` en réutilisant `registerMember(t)` ou
   `adminSession(t)`.
3. Lançez `make check` et vérifiez que la suite est verte.

## CI

Ces tests peuvent s'exécuter tels quels dans une GitHub Actions : l'étape de CI
n'a qu'à appeler `make test-integration`. Le conteneur PostgreSQL et l'API sont
gérés par `scripts/run.sh`, aucun service annexe n'est requis.

## Diagnostic

- Le serveur échoue au démarrage ? Consultez `/tmp/ojozone-test-api.log` : en
  cas d'échec, `run.sh` affiche automatiquement ses 50 dernières lignes.
- PostgreSQL reste actif après la suite : `make -C tests down` pour l'arrêter.
- La base `ojozone_test` persiste entre deux passes : utile en développement
  (les tests sont idempotents : emails aléatoires, upserts par clé fixe ).