-- name: UpsertGeoAreaByID :one
INSERT INTO geo_areas (id, parent_id, type, code, name, country_code, centroid)
VALUES (
    sqlc.arg('id')::uuid,
    sqlc.narg('parent_id')::uuid,
    sqlc.arg('type')::varchar,
    sqlc.narg('code')::varchar,
    sqlc.arg('name')::text,
        sqlc.arg('country_code')::char(2),
        CASE WHEN sqlc.narg('latitude')::double precision IS NULL OR sqlc.narg('longitude')::double precision IS NULL THEN NULL
            ELSE ST_SetSRID(ST_MakePoint(sqlc.narg('longitude')::double precision, sqlc.narg('latitude')::double precision), 4326)::geography END
)
ON CONFLICT (id) DO UPDATE SET
    parent_id = EXCLUDED.parent_id,
    type = EXCLUDED.type,
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    country_code = EXCLUDED.country_code,
    centroid = EXCLUDED.centroid
RETURNING *;

-- name: UpsertGeoAreaByCode :one
INSERT INTO geo_areas (parent_id, type, code, name, country_code, centroid)
VALUES (
    sqlc.narg('parent_id')::uuid,
    sqlc.arg('type')::varchar,
    sqlc.arg('code')::varchar,
    sqlc.arg('name')::text,
        sqlc.arg('country_code')::char(2),
        CASE WHEN sqlc.narg('latitude')::double precision IS NULL OR sqlc.narg('longitude')::double precision IS NULL THEN NULL
            ELSE ST_SetSRID(ST_MakePoint(sqlc.narg('longitude')::double precision, sqlc.narg('latitude')::double precision), 4326)::geography END
)
ON CONFLICT (type, country_code, code) DO UPDATE SET
    parent_id = EXCLUDED.parent_id,
    name = EXCLUDED.name,
    centroid = EXCLUDED.centroid
RETURNING *;

-- name: UpsertSourceByID :one
INSERT INTO sources (
    id, name, kind, homepage_url, license_name, license_url, attribution, is_active
) VALUES (
    sqlc.arg('id')::uuid,
    sqlc.arg('name')::text,
    sqlc.arg('kind')::source_kind,
    sqlc.narg('homepage_url')::text,
    sqlc.narg('license_name')::text,
    sqlc.narg('license_url')::text,
    sqlc.narg('attribution')::text,
    sqlc.arg('is_active')::boolean
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    kind = EXCLUDED.kind,
    homepage_url = EXCLUDED.homepage_url,
    license_name = EXCLUDED.license_name,
    license_url = EXCLUDED.license_url,
    attribution = EXCLUDED.attribution,
    is_active = EXCLUDED.is_active
RETURNING *;

-- name: UpsertCategoryBySlug :one
INSERT INTO categories (parent_id, slug, name_i18n)
VALUES (
    sqlc.narg('parent_id')::uuid,
    sqlc.arg('slug')::text,
    sqlc.arg('name_i18n')::jsonb
)
ON CONFLICT (slug) DO UPDATE SET
    parent_id = EXCLUDED.parent_id,
    name_i18n = EXCLUDED.name_i18n
RETURNING *;

-- name: UpsertUnitByCode :one
INSERT INTO units (code, dimension, to_base_factor)
VALUES (
    sqlc.arg('code')::varchar,
    sqlc.arg('dimension')::varchar,
    sqlc.arg('to_base_factor')::numeric
)
ON CONFLICT (code) DO UPDATE SET
    dimension = EXCLUDED.dimension,
    to_base_factor = EXCLUDED.to_base_factor
RETURNING *;

-- name: UpsertFuelTypeByCode :one
INSERT INTO fuel_types (code, name_i18n, energy)
VALUES (
    sqlc.arg('code')::varchar,
    sqlc.arg('name_i18n')::jsonb,
    sqlc.arg('energy')::varchar
)
ON CONFLICT (code) DO UPDATE SET
    name_i18n = EXCLUDED.name_i18n,
    energy = EXCLUDED.energy
RETURNING *;

-- name: UpsertMerchantByID :one
INSERT INTO merchants (id, name, website_url)
VALUES (
    sqlc.arg('id')::uuid,
    sqlc.arg('name')::text,
    sqlc.narg('website_url')::text
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    website_url = EXCLUDED.website_url
RETURNING *;

-- name: UpsertLocationByID :one
INSERT INTO locations (id, merchant_id, geo_area_id, name, address, position, external_ref)
VALUES (
    sqlc.arg('id')::uuid,
    sqlc.narg('merchant_id')::uuid,
    sqlc.arg('geo_area_id')::uuid,
    sqlc.arg('name')::text,
        sqlc.narg('address')::text,
        CASE WHEN sqlc.narg('latitude')::double precision IS NULL OR sqlc.narg('longitude')::double precision IS NULL THEN NULL
            ELSE ST_SetSRID(ST_MakePoint(sqlc.narg('longitude')::double precision, sqlc.narg('latitude')::double precision), 4326)::geography END,
    sqlc.narg('external_ref')::text
)
ON CONFLICT (id) DO UPDATE SET
    merchant_id = EXCLUDED.merchant_id,
    geo_area_id = EXCLUDED.geo_area_id,
    name = EXCLUDED.name,
    address = EXCLUDED.address,
    position = EXCLUDED.position,
    external_ref = EXCLUDED.external_ref
RETURNING *;
