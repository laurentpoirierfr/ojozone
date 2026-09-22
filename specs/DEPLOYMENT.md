# OjoZone - Hébergement de test

> Proposition vérifiée le 22 septembre 2026. Les offres gratuites peuvent évoluer ; leurs pages tarifaires restent la référence.

## 1. Décision retenue

OjoZone démarre sous la forme d'un **monolithe modulaire dans un seul conteneur Docker**. Ce conteneur sert l'interface web, l'API REST et l'administration. Il est construit et déployé directement par Render depuis le dépôt Git.

La première version utilise seulement deux services :

| Besoin | Service | Offre de test |
|---|---|---|
| Site, API et administration | [Render](https://render.com/docs/free) | Un Web Service Docker Free |
| Base PostgreSQL avec PostGIS | [Neon](https://neon.com/pricing) | Free |
| DNS de test | Render | Sous-domaine `*.onrender.com` et HTTPS inclus |

Cloudflare R2 sera ajouté dans un second temps, au moment d'implémenter l'envoi de photos. GitHub Container Registry, Redis, une file de messages et des workers séparés ne font pas partie du démarrage.

URL visée :

```text
https://ojozone.onrender.com
```

Le nom exact n'est **pas garanti** avant la création du service dans Render. Le DNS `onrender.com` est générique : une réponse DNS ne prouve donc pas que le nom est réservable. Si `ojozone` est refusé, utiliser par ordre de préférence :

```text
https://ojozone-app.onrender.com
https://ojozone-eu.onrender.com
https://ojozone-demo.onrender.com
```

Cette pile est adaptée au développement, aux tests fonctionnels et aux démonstrations, pas à une production avec engagement de disponibilité.

### Ce qui tourne dans le conteneur unique

- les pages web publiques ;
- les écrans de connexion et d'administration ;
- l'API sous `/api/v1` ;
- les contrôles simples et les agrégations déclenchées à la demande ;
- la route de santé `/health`.

Le conteneur ne contient ni PostgreSQL, ni données persistantes, ni fichiers envoyés par les utilisateurs.

## 2. Pourquoi cette solution

Render sait construire un `Dockerfile` depuis un dépôt Git ou exécuter une image Docker/OCI existante. Il fournit automatiquement une URL publique et un certificat TLS. Le service gratuit dispose actuellement de 512 Mo de RAM, de 0,1 CPU et d'un quota partagé de 750 heures par mois.

Neon est retenu à la place de la base gratuite Render, car cette dernière expire après 30 jours. L'offre Neon Free est permanente, se met en veille après cinq minutes d'inactivité et permet d'activer PostGIS, nécessaire au modèle géographique défini dans les spécifications.

Lorsque la contribution avec photo sera développée, Cloudflare R2 évitera de stocker les fichiers dans le conteneur Render, dont le système de fichiers est éphémère. Son offre gratuite Standard comprend actuellement 10 Go-mois, un million d'opérations d'écriture, dix millions d'opérations de lecture par mois et la sortie réseau gratuite.

## 3. Architecture de test

```mermaid
flowchart LR
	User[Navigateur et application mobile] -->|HTTPS| Render[Conteneur unique Render<br/>Web + API + administration]
	Render -->|TLS PostgreSQL| Neon[(Neon PostgreSQL + PostGIS)]
	GitHub[Dépôt Git avec Dockerfile] -->|Push : build et déploiement| Render
	Render -. Phase contribution .-> R2[(Cloudflare R2)]
```

Render construit directement le `Dockerfile` à chaque push sur la branche de déploiement. Les traitements courts restent dans le processus principal. Une tâche trop longue doit d'abord être exposée comme une commande d'administration contrôlée ; elle ne justifiera un worker séparé que lorsque sa durée ou son volume perturbera l'API.

### Éléments explicitement différés

- Redis : inutile tant que les requêtes PostgreSQL et le cache HTTP suffisent.
- Worker et file de messages : à ajouter uniquement pour l'OCR, les imports lourds ou les agrégations coûteuses.
- GHCR : à ajouter lorsqu'une même image doit être promue entre plusieurs environnements.
- Application mobile : elle consommera la même API publique sans modifier l'hébergement initial.

## 4. Limites connues

### Render Free

- Mise en veille après 15 minutes sans requête entrante.
- Redémarrage à froid pouvant prendre environ une minute.
- Système de fichiers local effacé lors d'un redéploiement, redémarrage ou passage en veille.
- Un seul port HTTP public ; l'application doit écouter sur `0.0.0.0:$PORT`.
- Pas de disque persistant, de montée en charge ni de garantie de disponibilité.
- L'épuisement des heures ou de la bande passante gratuite peut suspendre le service.
- Une image Docker préconstruite doit être redéployée manuellement ; le déploiement depuis un dépôt avec `Dockerfile` permet l'intégration continue native.

### Neon Free

- 0,5 Go de stockage par projet.
- 100 CU-heures de calcul et 5 Go de trafic sortant par projet et par mois.
- Mise en veille automatique après cinq minutes ; la première connexion peut être plus lente.
- Une limite atteinte bloque temporairement le calcul ou les écritures, mais ne supprime pas les données.

### Cloudflare R2

- Utiliser la classe `Standard`, seule classe couverte par le quota gratuit.
- Le domaine de test `r2.dev` est limité et n'est pas destiné à la production.
- Garder le bucket privé et fournir des URLs signées à courte durée pour les preuves.
- Une carte ou une vérification de paiement peut être demandée à l'ouverture du compte, même si l'usage reste dans le quota gratuit.

## 5. Nom de domaine

État constaté le 22 septembre 2026 :

| Nom | État observé | Décision |
|---|---|---|
| `ojozone.com` | Déjà enregistré depuis le 12 octobre 2025 | Ne pas utiliser sans accord de son titulaire |
| `ojozone.fr` | Aucune fiche trouvée dans le RDAP de l'AFNIC | Candidat prioritaire, à confirmer et acheter immédiatement chez un registrar |
| `ojozone.eu` | Vérification RDAP non concluante | À vérifier directement chez EURid ou un registrar |
| `ojozone.onrender.com` | Disponibilité non vérifiable par DNS | À tenter lors de la création du service Render |

Un domaine `ojozone.fr` n'est pas gratuit. Pour la phase de test, `ojozone.onrender.com` suffit et bénéficie automatiquement de HTTPS. Pour une bêta publique, réserver le domaine sans tarder puis configurer :

```text
ojozone.fr       CNAME ou ALIAS vers Render
www.ojozone.fr   CNAME vers le nom fourni par Render
api.ojozone.fr   CNAME vers le service API si celui-ci est séparé
```

L'absence actuelle de fiche RDAP pour `ojozone.fr` est un indice, pas une réservation. Seule la confirmation et la commande auprès d'un bureau d'enregistrement garantissent le nom.

## 6. Ordre de mise en œuvre

### Étape 1 - Créer le socle applicatif

Le dépôt doit commencer avec la structure logique suivante, adaptée ensuite au framework retenu :

```text
/
|-- Dockerfile
|-- .dockerignore
|-- src/
|   |-- web/
|   |-- api/
|   `-- modules/
|-- deploy/
|   `-- migration/
|       `-- sql/
`-- tests/
```

Les modules métier restent séparés dans le code, mais sont compilés et exécutés ensemble. Le frontend appelle l'API avec des chemins relatifs (`/api/v1`) afin d'éviter immédiatement CORS et plusieurs noms DNS.

### Étape 2 - Préparer le conteneur

Le conteneur doit :

- exposer un unique serveur HTTP ;
- écouter sur l'adresse `0.0.0.0` et le port fourni par `PORT` ;
- fournir une route `GET /health` sans dépendance lourde ;
- écrire les journaux sur la sortie standard ;
- ne conserver aucun fichier utilisateur sur son disque local ;
- exécuter les migrations SQL séparément et de façon idempotente.

Le `Dockerfile` doit utiliser une construction multi-stage, une image d'exécution minimale et un utilisateur non privilégié. Le serveur ne doit pas supposer que `PORT` possède une valeur fixe.

### Étape 3 - Créer la base Neon

1. Créer un projet dans une région européenne disponible.
2. Copier l'URL directe avec TLS pour les migrations et l'URL avec pooling pour l'application.
3. Exporter temporairement l'URL directe puis exécuter le provisionnement :

```sh
export DATABASE_URL='postgresql://USER:PASSWORD@HOST/DATABASE?sslmode=require&search_path=public'
./deploy/provision.sh
```

4. Vérifier que le script affiche la version `1` du schéma.
5. Stocker uniquement l'URL poolée dans le secret Render `DATABASE_URL`.

Les migrations versionnées se trouvent dans `deploy/migration/sql` et sont exécutées avec [`golang-migrate`](https://github.com/golang-migrate/migrate). Le détail des commandes est disponible dans [le guide de déploiement](../deploy/README.md).

### Étape 4 - Créer le service Render

1. Créer un **Blueprint** à partir de `deploy/render.yaml` et connecter le dépôt Git.
2. Choisir le runtime Docker et le plan Free.
3. Demander le nom `ojozone`.
4. Choisir Francfort et la même zone géographique que Neon lorsque cela est possible.
5. Ajouter `DATABASE_URL` dans les secrets Render.
6. Configurer `/health` comme chemin de contrôle.
7. Déployer, puis tester `https://ojozone.onrender.com/health`.
8. Activer le déploiement automatique uniquement depuis la branche principale.

### Étape 5 - Livrer une première verticale fonctionnelle

Avant d'ajouter un autre service d'infrastructure, livrer ce parcours complet :

1. charger les migrations et quelques données de référence ;
2. rechercher un produit avec `GET /api/v1/products` ;
3. consulter ses prix avec `GET /api/v1/products/{id}/prices` ;
4. afficher ces résultats dans l'interface web ;
5. vérifier les logs et la reprise après mise en veille.

### Étape 6 - Ajouter R2 avec la contribution photo

1. Créer un bucket privé `ojozone-evidence-dev` en classe Standard.
2. Créer des identifiants S3 limités à ce bucket.
3. Définir les secrets Render :

```text
S3_ENDPOINT
S3_REGION=auto
S3_BUCKET=ojozone-evidence-dev
S3_ACCESS_KEY_ID
S3_SECRET_ACCESS_KEY
```

4. Faire générer les URLs signées d'envoi et de lecture par l'API OjoZone.
5. Ajouter une règle de cycle de vie pour supprimer les preuves de test devenues inutiles.

## 7. Évolution optionnelle vers GHCR

Une image publique peut être stockée sous la forme :

```text
ghcr.io/<organisation>/ojozone:<version>
```

Le flux recommandé est :

1. GitHub Actions construit l'image à partir d'un commit ou d'un tag.
2. Le workflow s'authentifie avec `GITHUB_TOKEN` et pousse l'image dans GHCR.
3. Le paquet est rendu public pour que Render puisse le télécharger sans secret.
4. Render déploie une version immuable ou, de préférence, un digest `sha256`.

Ne jamais utiliser uniquement le tag `latest` pour une version à conserver : un digest rend le déploiement reproductible.

Cette évolution est différée : pour le premier déploiement, Render construit l'image depuis le `Dockerfile` du dépôt.

## 8. Variables d'environnement minimales

```text
APP_ENV=development
PUBLIC_BASE_URL=https://ojozone.onrender.com
DATABASE_URL=<secret Neon>
```

À ajouter seulement lorsque R2 est activé :

```text
S3_ENDPOINT=<secret ou configuration>
S3_REGION=auto
S3_BUCKET=ojozone-evidence-dev
S3_ACCESS_KEY_ID=<secret>
S3_SECRET_ACCESS_KEY=<secret>
```

Les clés, mots de passe et URLs contenant des identifiants restent exclusivement dans le gestionnaire de secrets Render.

## 9. Critères de validation

- L'image Docker démarre sans volume persistant.
- `/health` répond en HTTPS sur le sous-domaine Render.
- L'API se reconnecte après la mise en veille de Render et de Neon.
- `SELECT PostGIS_Version();` réussit sur Neon.
- Une photo est envoyée directement vers R2 par URL signée puis relue avec autorisation.
- Aucun secret n'apparaît dans les logs, l'image Docker ou le dépôt Git.
- Une migration SQL et un retour à l'image précédente ont été testés.
- Le budget reste à 0 EUR dans les tableaux de consommation Render et Neon, puis R2 lorsqu'il est activé.

## 10. Passage en production

La première dépense doit porter sur le calcul Render afin de supprimer la veille et d'obtenir davantage de ressources. Il faudra ensuite augmenter la capacité et les sauvegardes PostgreSQL, attacher le domaine `ojozone.fr`, configurer un CDN pour les médias et séparer le worker de traitement des photos de l'API publique.

Les quotas gratuits ne doivent contenir que des données de test anonymisées : ils n'offrent ni les garanties, ni les sauvegardes, ni les engagements nécessaires à des données utilisateurs réelles.

## 11. Sources officielles

- [Render - Offre gratuite](https://render.com/docs/free)
- [Render - Services web et Docker](https://render.com/docs/web-services)
- [Neon - Tarification](https://neon.com/pricing)
- [Neon - Extension PostGIS](https://neon.com/docs/extensions/postgis)
- [Cloudflare R2 - Tarification](https://developers.cloudflare.com/r2/pricing/)
- [Cloudflare R2 - Limites](https://developers.cloudflare.com/r2/platform/limits/)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
