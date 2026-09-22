BEGIN;

DROP TABLE IF EXISTS price_aggregates_monthly;
DROP TABLE IF EXISTS moderation_events;
DROP TABLE IF EXISTS evidence_files;
DROP TABLE IF EXISTS income_observations;
DROP TABLE IF EXISTS housing_observations;
DROP TABLE IF EXISTS fuel_price_observations;
DROP TABLE IF EXISTS fuel_types;
DROP TABLE IF EXISTS product_price_observations;
DROP TABLE IF EXISTS locations;
DROP TABLE IF EXISTS merchants;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS sources;
DROP TABLE IF EXISTS geo_areas;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS income_basis;
DROP TYPE IF EXISTS income_period;
DROP TYPE IF EXISTS housing_transaction;
DROP TYPE IF EXISTS source_kind;
DROP TYPE IF EXISTS moderation_status;
DROP TYPE IF EXISTS user_role;

COMMIT;
