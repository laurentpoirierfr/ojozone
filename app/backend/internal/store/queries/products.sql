-- name: ListProducts :many
SELECT
    p.id,
    p.name,
    p.brand,
    p.barcode,
    p.reference_unit_code,
    p.is_generic,
    p.attributes,
    p.created_at,
    c.slug AS category_slug,
    c.name_i18n AS category_name_i18n
FROM products AS p
JOIN categories AS c ON c.id = p.category_id
WHERE (
    sqlc.narg('search')::text IS NULL
    OR p.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR p.brand ILIKE '%' || sqlc.narg('search')::text || '%'
    OR p.barcode = sqlc.narg('search')::text
)
ORDER BY p.name, p.id
LIMIT sqlc.arg('page_size')::integer
OFFSET sqlc.arg('page_offset')::integer;

-- name: GetProduct :one
SELECT
    p.id,
    p.name,
    p.brand,
    p.barcode,
    p.reference_unit_code,
    p.is_generic,
    p.attributes,
    p.created_at,
    c.slug AS category_slug,
    c.name_i18n AS category_name_i18n
FROM products AS p
JOIN categories AS c ON c.id = p.category_id
WHERE p.id = sqlc.arg('id')::uuid;

-- name: GetProductByBarcode :one
SELECT
    p.id,
    p.name,
    p.brand,
    p.barcode,
    p.reference_unit_code,
    p.is_generic,
    p.attributes,
    p.created_at,
    c.slug AS category_slug,
    c.name_i18n AS category_name_i18n
FROM products AS p
JOIN categories AS c ON c.id = p.category_id
WHERE p.barcode = sqlc.arg('barcode')::text;

-- name: UpsertProductByID :one
INSERT INTO products (
    id,
    category_id,
    name,
    brand,
    barcode,
    reference_unit_code,
    is_generic,
    attributes
) VALUES (
    sqlc.arg('id')::uuid,
    sqlc.arg('category_id')::uuid,
    sqlc.arg('name')::text,
    sqlc.narg('brand')::text,
    sqlc.narg('barcode')::text,
    sqlc.arg('reference_unit_code')::varchar,
    sqlc.arg('is_generic')::boolean,
    sqlc.arg('attributes')::jsonb
)
ON CONFLICT (id) DO UPDATE SET
    category_id = EXCLUDED.category_id,
    name = EXCLUDED.name,
    brand = EXCLUDED.brand,
    barcode = EXCLUDED.barcode,
    reference_unit_code = EXCLUDED.reference_unit_code,
    is_generic = EXCLUDED.is_generic,
    attributes = EXCLUDED.attributes
RETURNING *;

-- name: UpsertProductByBarcode :one
INSERT INTO products (
    category_id,
    name,
    brand,
    barcode,
    reference_unit_code,
    is_generic,
    attributes
) VALUES (
    sqlc.arg('category_id')::uuid,
    sqlc.arg('name')::text,
    sqlc.narg('brand')::text,
    sqlc.arg('barcode')::varchar,
    sqlc.arg('reference_unit_code')::varchar,
    sqlc.arg('is_generic')::boolean,
    sqlc.arg('attributes')::jsonb
)
ON CONFLICT (barcode) WHERE barcode IS NOT NULL DO UPDATE SET
    category_id = EXCLUDED.category_id,
    name = EXCLUDED.name,
    brand = EXCLUDED.brand,
    reference_unit_code = EXCLUDED.reference_unit_code,
    is_generic = EXCLUDED.is_generic,
    attributes = EXCLUDED.attributes
RETURNING *;

-- name: UpsertProductPriceBySourceRecord :one
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
    observed_at,
    status,
    confidence_score,
    source_record_id
) VALUES (
    sqlc.arg('product_id')::uuid,
    sqlc.arg('location_id')::uuid,
    sqlc.arg('source_id')::uuid,
    sqlc.narg('contributor_id')::uuid,
    sqlc.arg('amount')::numeric,
    sqlc.arg('currency')::char(3),
    sqlc.arg('quantity')::numeric,
    sqlc.arg('unit_code')::varchar,
    sqlc.arg('normalized_amount')::numeric,
    sqlc.arg('is_promotion')::boolean,
    sqlc.arg('observed_at')::timestamptz,
    sqlc.arg('status')::moderation_status,
    sqlc.narg('confidence_score')::numeric,
    sqlc.arg('source_record_id')::text
)
ON CONFLICT (source_id, source_record_id) WHERE source_record_id IS NOT NULL DO UPDATE SET
    product_id = EXCLUDED.product_id,
    location_id = EXCLUDED.location_id,
    contributor_id = EXCLUDED.contributor_id,
    amount = EXCLUDED.amount,
    currency = EXCLUDED.currency,
    quantity = EXCLUDED.quantity,
    unit_code = EXCLUDED.unit_code,
    normalized_amount = EXCLUDED.normalized_amount,
    is_promotion = EXCLUDED.is_promotion,
    observed_at = EXCLUDED.observed_at,
    status = EXCLUDED.status,
    confidence_score = EXCLUDED.confidence_score
RETURNING *;

-- name: UpsertProductPriceByID :one
INSERT INTO product_price_observations (
    id,
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
    observed_at,
    status,
    confidence_score,
    source_record_id
) VALUES (
    sqlc.arg('id')::uuid,
    sqlc.arg('product_id')::uuid,
    sqlc.arg('location_id')::uuid,
    sqlc.arg('source_id')::uuid,
    sqlc.narg('contributor_id')::uuid,
    sqlc.arg('amount')::numeric,
    sqlc.arg('currency')::char(3),
    sqlc.arg('quantity')::numeric,
    sqlc.arg('unit_code')::varchar,
    sqlc.arg('normalized_amount')::numeric,
    sqlc.arg('is_promotion')::boolean,
    sqlc.arg('observed_at')::timestamptz,
    sqlc.arg('status')::moderation_status,
    sqlc.narg('confidence_score')::numeric,
    sqlc.narg('source_record_id')::text
)
ON CONFLICT (id) DO UPDATE SET
    product_id = EXCLUDED.product_id,
    location_id = EXCLUDED.location_id,
    source_id = EXCLUDED.source_id,
    contributor_id = EXCLUDED.contributor_id,
    amount = EXCLUDED.amount,
    currency = EXCLUDED.currency,
    quantity = EXCLUDED.quantity,
    unit_code = EXCLUDED.unit_code,
    normalized_amount = EXCLUDED.normalized_amount,
    is_promotion = EXCLUDED.is_promotion,
    observed_at = EXCLUDED.observed_at,
    status = EXCLUDED.status,
    confidence_score = EXCLUDED.confidence_score,
    source_record_id = EXCLUDED.source_record_id
RETURNING *;

-- name: GetProductPrice :one
SELECT
    o.id,
    o.product_id,
    o.amount,
    o.currency,
    o.quantity,
    o.unit_code,
    o.normalized_amount,
    o.is_promotion,
    o.observed_at,
    o.status,
    o.confidence_score,
    o.source_record_id,
    l.id AS location_id,
    l.name AS location_name,
    l.address AS location_address,
    g.id AS geo_area_id,
    g.name AS geo_area_name,
    s.id AS source_id,
    s.name AS source_name,
    s.kind AS source_kind
FROM product_price_observations AS o
JOIN locations AS l ON l.id = o.location_id
JOIN geo_areas AS g ON g.id = l.geo_area_id
JOIN sources AS s ON s.id = o.source_id
WHERE o.id = sqlc.arg('id')::uuid;

-- name: DeleteProductPrice :execrows
DELETE FROM product_price_observations
WHERE id = sqlc.arg('id')::uuid;

-- name: DeleteProduct :execrows
DELETE FROM products
WHERE id = sqlc.arg('id')::uuid;

-- name: ListApprovedProductPrices :many
SELECT
    o.id,
    o.product_id,
    o.amount,
    o.currency,
    o.quantity,
    o.unit_code,
    o.normalized_amount,
    o.is_promotion,
    o.observed_at,
    o.status,
    o.confidence_score,
    o.source_record_id,
    l.id AS location_id,
    l.name AS location_name,
    l.address AS location_address,
    g.id AS geo_area_id,
    g.name AS geo_area_name,
    s.id AS source_id,
    s.name AS source_name,
    s.kind AS source_kind
FROM product_price_observations AS o
JOIN locations AS l ON l.id = o.location_id
JOIN geo_areas AS g ON g.id = l.geo_area_id
JOIN sources AS s ON s.id = o.source_id
WHERE o.product_id = sqlc.arg('product_id')::uuid
  AND o.status = 'approved'
  AND (sqlc.narg('geo_area_id')::uuid IS NULL OR g.id = sqlc.narg('geo_area_id')::uuid)
ORDER BY o.observed_at DESC, o.id
LIMIT sqlc.arg('page_size')::integer
OFFSET sqlc.arg('page_offset')::integer;

-- name: ListProductPriceObservations :many
SELECT
        o.id,
        o.product_id,
        o.amount,
        o.currency,
        o.quantity,
        o.unit_code,
        o.normalized_amount,
        o.is_promotion,
        o.observed_at,
        o.status,
        o.confidence_score,
        o.source_record_id,
        l.id AS location_id,
        l.name AS location_name,
        l.address AS location_address,
        g.id AS geo_area_id,
        g.name AS geo_area_name,
        s.id AS source_id,
        s.name AS source_name,
        s.kind AS source_kind
FROM product_price_observations AS o
JOIN locations AS l ON l.id = o.location_id
JOIN geo_areas AS g ON g.id = l.geo_area_id
JOIN sources AS s ON s.id = o.source_id
WHERE (sqlc.narg('product_id')::uuid IS NULL OR o.product_id = sqlc.narg('product_id')::uuid)
    AND (sqlc.narg('geo_area_id')::uuid IS NULL OR g.id = sqlc.narg('geo_area_id')::uuid)
    AND (sqlc.narg('status')::text IS NULL OR o.status::text = sqlc.narg('status')::text)
ORDER BY o.observed_at DESC, o.id
LIMIT sqlc.arg('page_size')::integer
OFFSET sqlc.arg('page_offset')::integer;
