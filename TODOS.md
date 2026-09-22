# OjoZone - Suivi du projet

> Dernière mise à jour : 22 septembre 2026  
> Légende : `[x]` réalisé, `[~]` partiellement réalisé, `[ ]` à faire

## Vue d'ensemble

| Domaine | État | Prochaine étape |
|---|---|---|
| Spécifications | Avancé | Valider le périmètre pilote et les sources de données |
| Base de données | Socle opérationnel | Ajouter données de référence, évolutions et tests d'intégration |
| Backend/API | Socle avancé | Sécuriser, fiabiliser et compléter les parcours métier |
| Frontend web | Non démarré | Initialiser l'application et livrer la recherche produit |
| Application mobile | Non démarré | Choisir la technologie après stabilisation de l'API |
| Déploiement | Développement prêt | Ajouter le Dockerfile applicatif et déployer sur Render/Neon |
| Qualité/CI | Partiel | Créer la CI et élargir les tests d'intégration |
| Données | Non démarré | Choisir les villes pilotes et importer les premières sources |

## Priorités immédiates

### P0 - Première démonstration fonctionnelle

- [ ] Choisir les villes et pays pilotes.
- [ ] Définir le panier initial de produits essentiels et les unités canoniques.
- [ ] Ajouter une migration ou un mécanisme de seed pour les unités, catégories, sources et zones pilotes.
- [ ] Initialiser le frontend web dans `app/frontend`.
- [ ] Implémenter la recherche et la liste des produits.
- [ ] Implémenter la fiche produit avec ses prix récents.
- [ ] Ajouter le `Dockerfile` multi-stage servant le frontend compilé avec le backend Go.
- [ ] Ajouter le backend au `docker-compose.yaml`.
- [ ] Déployer une première version sur Render avec la base Neon.
- [ ] Vérifier le parcours complet : navigateur → API → PostgreSQL.

### P1 - Contributions et sécurité

- [ ] Mettre en place l'authentification OpenID Connect/OAuth 2.1.
- [ ] Appliquer les rôles membre, modérateur, administrateur et partenaire.
- [ ] Protéger toutes les écritures et les routes `/api/v1/admin/*`.
- [ ] Ajouter les contributions produit et carburant avec statut `pending`.
- [ ] Ajouter Cloudflare R2 et les URLs signées pour les preuves photo.
- [ ] Implémenter la file de modération.
- [ ] Ajouter quotas, limitation de débit et protection anti-abus.

### P2 - Données et expérience complète

- [ ] Automatiser les imports de données officielles.
- [ ] Implémenter les comparaisons entre villes et pays.
- [ ] Calculer les agrégats mensuels et les scores de confiance.
- [ ] Ajouter immobilier, revenus et carburants dans l'interface web.
- [ ] Concevoir puis développer l'application mobile.
- [ ] Ajouter l'internationalisation française et anglaise.

## Spécifications

### Réalisé

- [x] Définir la vision, les utilisateurs et les rôles.
- [x] Décrire le périmètre MVP et les évolutions ultérieures.
- [x] Décrire les parcours de comparaison, contribution et import.
- [x] Définir les règles métier principales.
- [x] Proposer l'architecture cible.
- [x] Définir le schéma PostgreSQL/PostGIS initial.
- [x] Décrire les conventions et les routes REST proposées.
- [x] Définir les principes de qualité, modération, sécurité et conformité.
- [x] Définir les exigences non fonctionnelles et critères d'acceptation.
- [x] Documenter la stratégie de déploiement de test.

Documents :

- `specs/README.md`
- `specs/DEPLOYMENT.md`

### À valider

- [ ] Arrêter la liste des cinq villes pilotes dans au moins trois pays.
- [ ] Confirmer les sources officielles, leurs licences et leurs fréquences de mise à jour.
- [ ] Valider la composition du panier de référence.
- [ ] Valider les règles de normalisation des quantités et unités.
- [ ] Définir précisément le calcul des médianes, agrégats et scores de confiance.
- [ ] Définir le niveau de preuve requis pour publier une contribution.
- [ ] Définir les durées de conservation des photos, imports et journaux.
- [ ] Décider si `ojozone.fr` doit être réservé.

## Base de données et migrations

### Réalisé

- [x] Créer le schéma PostgreSQL avec l'extension PostGIS.
- [x] Modéliser utilisateurs, géographie, sources, catalogue et unités.
- [x] Modéliser produits, prix, carburants, immobilier et revenus.
- [x] Modéliser preuves, modération et agrégats mensuels.
- [x] Créer les index géographiques, de recherche et d'unicité.
- [x] Créer les migrations `up` et `down` compatibles avec `golang-migrate`.
- [x] Ajouter les scripts de provisionnement et de création de migrations.
- [x] Valider un cycle réel `up → down → up` sur PostgreSQL 16/PostGIS.
- [x] Ajouter PostgreSQL/PostGIS et `golang-migrate` au Compose local.

### À faire

- [ ] Créer des seeds versionnés pour les unités, catégories et zones pilotes.
- [ ] Ajouter des tests automatiques des contraintes, index uniques et suppressions référencées.
- [ ] Définir une stratégie de sauvegarde et restauration Neon.
- [ ] Définir la rétention et le partitionnement éventuel des observations volumineuses.
- [ ] Ajouter une migration pour chaque évolution future ; ne jamais modifier la migration initiale après mise en service.

## Backend Go

### Réalisé

- [x] Initialiser le module Go dans `app/backend`.
- [x] Mettre en place Gin et pgx/v5.
- [x] Configurer sqlc.dev avec PostgreSQL.
- [x] Générer les accès typés depuis les requêtes SQL.
- [x] Séparer les couches `handler → service → repository → sqlc → PostgreSQL`.
- [x] Servir les fichiers de `app/backend/static` avec fallback SPA.
- [x] Ajouter l'arrêt propre du serveur HTTP.
- [x] Exposer `/ops/liveness`, `/ops/readiness` et `/ops/infos`.
- [x] Exposer les produits et leurs observations de prix.
- [x] Implémenter les upserts produit par UUID ou code-barres.
- [x] Implémenter les upserts de prix par source et identifiant externe.
- [x] Exposer les référentiels : zones, sources, catégories, unités, commerçants, lieux et types de carburant.
- [x] Exposer les observations : carburants, immobilier et revenus.
- [x] Exposer les preuves et les agrégats mensuels.
- [x] Exposer la modération en mode append-only.
- [x] Exposer les utilisateurs administratifs en lecture seule sans `password_hash`.
- [x] Standardiser les erreurs HTTP avec `application/problem+json`.
- [x] Ajouter Swagger/Swaggo et générer les spécifications JSON/YAML.
- [x] Ajouter le Makefile backend : génération, formatage, analyse, tests, build et exécution.

Documentation interactive : `http://localhost:8080/swagger/index.html`

### Partiellement réalisé

- [~] Validation métier : règles principales présentes, mais couverture à compléter pour toutes les ressources génériques.
- [~] Tests unitaires : handlers et règles d'upsert principales couverts, mais pas chaque ressource.
- [~] Tests d'intégration : scénarios manuels validés contre PostgreSQL, mais non automatisés en CI.
- [~] Pagination : `limit/offset` disponible ; le curseur prévu dans les spécifications reste à implémenter.
- [~] API d'administration : routes présentes mais non protégées par authentification/autorisation.

### À faire

- [ ] Ajouter l'authentification et l'autorisation par rôle.
- [ ] Séparer clairement les routes publiques, membre, partenaire et administration.
- [ ] Ajouter des DTO et validations dédiés à chaque ressource au lieu de réponses génériques lorsque nécessaire.
- [ ] Documenter dans Swagger tous les paramètres de chemin des ressources génériques.
- [ ] Ajouter des filtres métier : période, zone, statut, source et fraîcheur.
- [ ] Implémenter la pagination par curseur.
- [ ] Ajouter les routes de comparaison et d'historique prévues par les spécifications.
- [ ] Implémenter les imports CSV/JSON avec rapport d'erreurs et idempotence.
- [ ] Ajouter le dépôt et la lecture de preuves via URLs R2 signées.
- [ ] Ajouter la détection de doublons et de valeurs aberrantes.
- [ ] Ajouter les règles de transition des statuts de modération.
- [ ] Ajouter métriques, traces et journaux corrélés.
- [ ] Ajouter les tests de concurrence des upserts.
- [ ] Ajouter les tests de sécurité et de contrôle d'accès.

## API exposée

### Opérations

- [x] `GET /ops/liveness`
- [x] `GET /ops/readiness`
- [x] `GET /ops/infos`
- [x] Swagger UI sous `/swagger/index.html`

### Catalogue et référentiels

- [x] Produits : lecture, recherche, upsert, remplacement et suppression.
- [x] Prix produit : liste filtrée, lecture, upsert, remplacement et suppression.
- [x] Zones géographiques : CRUD/upsert.
- [x] Sources : CRUD/upsert.
- [x] Catégories : CRUD/upsert.
- [x] Unités : CRUD/upsert par code.
- [x] Commerçants : CRUD/upsert.
- [x] Lieux : CRUD/upsert.
- [x] Types de carburant : CRUD/upsert.

### Observations et administration

- [x] Prix carburant : CRUD/upsert.
- [x] Observations immobilières : CRUD/upsert.
- [x] Observations de revenus : CRUD/upsert.
- [x] Fichiers de preuve : métadonnées CRUD/upsert.
- [x] Agrégats mensuels : CRUD/upsert par clé composite.
- [x] Événements de modération : liste, lecture et ajout append-only.
- [x] Utilisateurs administratifs : liste et lecture sans secret.

### Manquant dans l'API métier

- [ ] Authentification : inscription, connexion, renouvellement et révocation.
- [ ] Profil courant : `GET/PATCH/DELETE /api/v1/me`.
- [ ] Contributions de l'utilisateur courant et suivi de statut.
- [ ] Signalement d'une observation.
- [ ] Recherche géographique de proximité.
- [ ] Historique agrégé produits, carburants, immobilier et revenus.
- [ ] Comparaison entre plusieurs zones et panier personnalisé.
- [ ] Import, validation et publication de fichiers.
- [ ] Fusion de produits en doublon.

## Frontend web

### Réalisé

- [x] Créer une page HTML statique minimale servie par le backend.

### À faire

- [ ] Choisir et initialiser le framework frontend dans `app/frontend`.
- [ ] Configurer le build vers `app/backend/static`.
- [ ] Créer le système visuel responsive et accessible.
- [ ] Ajouter navigation, recherche globale et gestion des erreurs.
- [ ] Créer la page de recherche et la fiche produit.
- [ ] Créer la comparaison de villes et pays.
- [ ] Créer les vues carburant, immobilier et revenus.
- [ ] Créer les formulaires de contribution.
- [ ] Créer l'espace utilisateur et le suivi des contributions.
- [ ] Créer l'espace de modération et d'administration.
- [ ] Ajouter cartes, graphiques historiques et indicateurs de confiance.
- [ ] Ajouter français/anglais, accessibilité WCAG 2.2 AA et SEO.
- [ ] Ajouter tests unitaires, composants et parcours navigateur.

## Application mobile

### À faire

- [ ] Choisir la technologie multiplateforme.
- [ ] Définir les parcours mobiles et la navigation.
- [ ] Implémenter authentification et stockage sécurisé des jetons.
- [ ] Implémenter consultation, recherche et comparaison.
- [ ] Implémenter contribution de prix et sélection du point de vente.
- [ ] Ajouter appareil photo, compression et dépôt de preuve.
- [ ] Ajouter scan de code-barres.
- [ ] Ajouter saisie hors ligne et synchronisation différée.
- [ ] Publier les versions de test iOS et Android.

## Déploiement et exploitation

### Réalisé

- [x] Documenter la pile gratuite Render + Neon + Cloudflare R2.
- [x] Fournir un Blueprint Render dans `deploy/render.yaml`.
- [x] Fournir les variables d'environnement d'exemple.
- [x] Fournir le provisionnement PostgreSQL avec `golang-migrate`.
- [x] Fournir un environnement Compose PostgreSQL/PostGIS local.
- [x] Ajouter les commandes Make de démarrage, migration, logs et reset.
- [x] Générer le modèle du domaine en PlantUML et PNG dans `assets/`.

### À faire

- [ ] Ajouter un `Dockerfile` multi-stage pour le backend et le frontend statique.
- [ ] Ajouter le service backend au Compose local.
- [ ] Aligner le healthcheck Render sur `/ops/readiness`.
- [ ] Provisionner réellement Neon et enregistrer les secrets dans Render.
- [ ] Déployer l'application sur Render et confirmer le sous-domaine.
- [ ] Réserver et connecter le domaine public retenu.
- [ ] Créer le bucket R2 privé et sa politique de cycle de vie.
- [ ] Ajouter sauvegardes, restauration testée et procédure de reprise.
- [ ] Ajouter supervision, alertes et suivi des quotas gratuits.
- [ ] Préparer un environnement de préproduction distinct.

## Qualité, sécurité et CI/CD

### Réalisé

- [x] Ajouter `go fmt`, `go vet`, tests et build dans le Makefile backend.
- [x] Ajouter des tests HTTP des routes opérationnelles, statiques et métier.
- [x] Ajouter des tests de service sur les upserts produit et prix.
- [x] Vérifier les migrations sur PostgreSQL/PostGIS réel.
- [x] Vérifier plusieurs parcours CRUD/upsert contre la base locale.
- [x] Protéger les secrets locaux avec `.gitignore`.

### À faire

- [ ] Ajouter une CI GitHub Actions : génération, diff propre, lint, tests et build.
- [ ] Ajouter des tests d'intégration reproductibles avec PostgreSQL/PostGIS en conteneur.
- [ ] Vérifier automatiquement que sqlc et Swagger sont à jour.
- [ ] Ajouter un linter Go et une analyse de vulnérabilités des dépendances.
- [ ] Ajouter tests de charge, limites de payload et timeouts par opération.
- [ ] Ajouter analyse antivirus et suppression EXIF des preuves.
- [ ] Définir CORS, CSP et en-têtes de sécurité pour la production.
- [ ] Éviter toute donnée personnelle réelle dans les environnements gratuits.

## Données et lancement pilote

### À faire

- [ ] Identifier et documenter une source pour chaque famille d'indicateurs.
- [ ] Créer un premier import reproductible de données officielles.
- [ ] Mesurer couverture, fraîcheur, taille d'échantillon et confiance.
- [ ] Publier une méthodologie compréhensible par les utilisateurs.
- [ ] Préparer un jeu de démonstration sans données personnelles.
- [ ] Organiser une bêta limitée et recueillir les retours.
- [ ] Définir les indicateurs de succès et le tableau de suivi.

## Définition de terminé du MVP

- [ ] Un visiteur recherche un produit et consulte ses prix récents.
- [ ] Un visiteur compare au moins deux villes sur les quatre familles d'indicateurs.
- [ ] Chaque valeur expose source, date, unité, échantillon et confiance.
- [ ] Un membre authentifié soumet un prix depuis un mobile.
- [ ] Une preuve photo est stockée de manière privée et temporaire.
- [ ] Un modérateur approuve ou rejette une contribution avec audit.
- [ ] Une contribution approuvée alimente les agrégats.
- [ ] Un administrateur importe un fichier sans créer de doublons.
- [ ] La suppression d'un compte anonymise les contributions conservées.
- [ ] Le web satisfait les exigences de performance et d'accessibilité définies.
- [ ] L'application est déployée avec sauvegarde, supervision et procédure de restauration.
