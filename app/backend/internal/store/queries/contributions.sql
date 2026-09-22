-- name: GetCommunitySourceID :one
SELECT id
FROM sources
WHERE kind = 'community'
ORDER BY created_at ASC, id
LIMIT 1;

-- name: InsertProductContribution :one
INSERT INTO product_price_observations (
    product_id,
    location_id,
    source_id,
    contributor_id,
    amount,
    currency,
    quantity,
    unit_code,
    normalized_amount,
    is_promotion,
    observed_at
) VALUES (
    sqlc.arg('product_id')::uuid,
    sqlc.arg('location_id')::uuid,
    sqlc.arg('source_id')::uuid,
    sqlc.arg('contributor_id')::uuid,
    sqlc.arg('amount')::numeric,
    sqlc.arg('currency')::char(3),
    sqlc.arg('quantity')::numeric,
    sqlc.arg('unit_code')::varchar,
    sqlc.arg('normalized_amount')::numeric,
    sqlc.arg('is_promotion')::boolean,
    sqlc.arg('observed_at')::timestamptz
)
RETURNING id, status;

-- name: InsertFuelContribution :one
INSERT INTO fuel_price_observations (
    fuel_type_id,
    location_id,
    source_id,
    contributor_id,
    amount_per_litre,
    currency,
    observed_at
) VALUES (
    sqlc.arg('fuel_type_id')::uuid,
    sqlc.arg('location_id')::uuid,
    sqlc.arg('source_id')::uuid,
    sqlc.arg('contributor_id')::uuid,
    sqlc.arg('amount_per_litre')::numeric,
    sqlc.arg('currency')::char(3),
    sqlc.arg('observed_at')::timestamptz
)
RETURNING id, status;

-- name: GetProductContribution :one
SELECT
    o.id,
    o.contributor_id,
    o.amount,
    o.currency,
    o.quantity,
    o.unit_code,
    o.is_promotion,
    o.observed_at,
    o.status,
    o.confidence_score,
    o.created_at,
    p.name AS product_name,
    l.id AS location_id,
    l.name AS location_name,
    g.id AS geo_area_id,
    g.name AS geo_area_name,
    s.id AS source_id,
    s.name AS source_name,
    s.kind AS source_kind
FROM product_price_observations AS o
JOIN products AS p ON p.id = o.product_id
JOIN locations AS l ON l.id = o.location_id
JOIN geo_areas AS g ON g.id = l.geo_area_id
JOIN sources AS s ON s.id = o.source_id
WHERE o.id = sqlc.arg('id')::uuid;

-- name: GetFuelContribution :one
SELECT
    o.id,
    o.contributor_id,
    o.amount_per_litre,
    o.currency,
    o.observed_at,
    o.status,
    o.confidence_score,
    o.created_at,
    f.code AS fuel_type_code,
    f.name_i18n AS fuel_type_name_i18n,
    l.id AS location_id,
    l.name AS location_name,
    g.id AS geo_area_id,
    g.name AS geo_area_name,
    s.id AS source_id,
    s.name AS source_name,
    s.kind AS source_kind
FROM fuel_price_observations AS o
JOIN fuel_types AS f ON f.id = o.fuel_type_id
JOIN locations AS l ON l.id = o.location_id
JOIN geo_areas AS g ON g.id = l.geo_area_id
JOIN sources AS s ON s.id = o.source_id
WHERE o.id = sqlc.arg('id')::uuid;

-- name: UpdateProductContribution :one
UPDATE product_price_observations
SET amount = sqlc.arg('amount')::numeric,
    currency = sqlc.arg('currency')::char(3),
    quantity = sqlc.arg('quantity')::numeric,
    unit_code = sqlc.arg('unit_code')::varchar,
    normalized_amount = sqlc.arg('normalized_amount')::numeric,
    is_promotion = sqlc.arg('is_promotion')::boolean,
    observed_at = sqlc.arg('observed_at')::timestamptz
WHERE id = sqlc.arg('id')::uuid
  AND contributor_id = sqlc.arg('contributor_id')::uuid
  AND status = 'pending'
RETURNING id;

-- name: UpdateFuelContribution :one
UPDATE fuel_price_observations
SET amount_per_litre = sqlc.arg('amount_per_litre')::numeric,
    currency = sqlc.arg('currency')::char(3),
    observed_at = sqlc.arg('observed_at')::timestamptz
WHERE id = sqlc.arg('id')::uuid
  AND contributor_id = sqlc.arg('contributor_id')::uuid
  AND status = 'pending'
RETURNING id;

-- name: DeleteProductContribution :execrows
DELETE FROM product_price_observations
WHERE id = sqlc.arg('id')::uuid
  AND contributor_id = sqlc.arg('contributor_id')::uuid
  AND status = 'pending';

-- name: DeleteFuelContribution :execrows
DELETE FROM fuel_price_observations
WHERE id = sqlc.arg('id')::uuid
  AND contributor_id = sqlc.arg('contributor_id')::uuid
  AND status = 'pending';