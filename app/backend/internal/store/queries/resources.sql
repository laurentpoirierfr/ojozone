-- name: ListGeoAreas :many
SELECT id, parent_id, type, code, name, country_code,
    COALESCE(ST_Y(centroid::geometry), 'NaN'::double precision) AS latitude,
    COALESCE(ST_X(centroid::geometry), 'NaN'::double precision) AS longitude
FROM geo_areas ORDER BY name, id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;

-- name: GetGeoArea :one
SELECT id, parent_id, type, code, name, country_code,
    COALESCE(ST_Y(centroid::geometry), 'NaN'::double precision) AS latitude,
    COALESCE(ST_X(centroid::geometry), 'NaN'::double precision) AS longitude
FROM geo_areas WHERE id = sqlc.arg('id')::uuid;

-- name: DeleteGeoArea :execrows
DELETE FROM geo_areas WHERE id=sqlc.arg('id')::uuid;

-- name: ListSources :many
SELECT * FROM sources ORDER BY name, id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetSource :one
SELECT * FROM sources WHERE id=sqlc.arg('id')::uuid;
-- name: DeleteSource :execrows
DELETE FROM sources WHERE id=sqlc.arg('id')::uuid;

-- name: ListCategories :many
SELECT * FROM categories ORDER BY slug LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetCategory :one
SELECT * FROM categories WHERE id=sqlc.arg('id')::uuid;
-- name: UpsertCategoryByID :one
INSERT INTO categories(id,parent_id,slug,name_i18n) VALUES(sqlc.arg('id')::uuid,sqlc.narg('parent_id')::uuid,sqlc.arg('slug')::text,sqlc.arg('name_i18n')::jsonb)
ON CONFLICT(id) DO UPDATE SET parent_id=EXCLUDED.parent_id,slug=EXCLUDED.slug,name_i18n=EXCLUDED.name_i18n RETURNING id;
-- name: DeleteCategory :execrows
DELETE FROM categories WHERE id=sqlc.arg('id')::uuid;

-- name: ListUnits :many
SELECT * FROM units ORDER BY code LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetUnit :one
SELECT * FROM units WHERE code=sqlc.arg('code')::varchar;
-- name: DeleteUnit :execrows
DELETE FROM units WHERE code=sqlc.arg('code')::varchar;

-- name: ListMerchants :many
SELECT * FROM merchants ORDER BY name,id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetMerchant :one
SELECT * FROM merchants WHERE id=sqlc.arg('id')::uuid;
-- name: DeleteMerchant :execrows
DELETE FROM merchants WHERE id=sqlc.arg('id')::uuid;

-- name: ListLocations :many
SELECT id,merchant_id,geo_area_id,name,address,external_ref,created_at,
COALESCE(ST_Y(position::geometry), 'NaN'::double precision) AS latitude,
COALESCE(ST_X(position::geometry), 'NaN'::double precision) AS longitude
FROM locations ORDER BY name,id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetLocation :one
SELECT id,merchant_id,geo_area_id,name,address,external_ref,created_at,
COALESCE(ST_Y(position::geometry), 'NaN'::double precision) AS latitude,
COALESCE(ST_X(position::geometry), 'NaN'::double precision) AS longitude
FROM locations WHERE id=sqlc.arg('id')::uuid;
-- name: DeleteLocation :execrows
DELETE FROM locations WHERE id=sqlc.arg('id')::uuid;

-- name: ListFuelTypes :many
SELECT * FROM fuel_types ORDER BY code LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetFuelType :one
SELECT * FROM fuel_types WHERE id=sqlc.arg('id')::uuid;
-- name: UpsertFuelTypeByID :one
INSERT INTO fuel_types(id,code,name_i18n,energy) VALUES(sqlc.arg('id')::uuid,sqlc.arg('code')::varchar,sqlc.arg('name_i18n')::jsonb,sqlc.arg('energy')::varchar)
ON CONFLICT(id) DO UPDATE SET code=EXCLUDED.code,name_i18n=EXCLUDED.name_i18n,energy=EXCLUDED.energy RETURNING id;
-- name: DeleteFuelType :execrows
DELETE FROM fuel_types WHERE id=sqlc.arg('id')::uuid;