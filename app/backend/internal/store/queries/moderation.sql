-- name: ListModerationQueue :many
SELECT
    id,
    observation_type,
    status,
    subject,
    contributor_id,
    contributor_email,
    amount,
    currency,
    quantity,
    unit_code,
    is_promotion,
    location_id,
    location_name,
    geo_area_id,
    geo_area_name,
    source_id,
    source_name,
    source_kind,
    observed_at,
    confidence_score,
    created_at
FROM (
    SELECT
        o.id AS id,
        'fuel_price'::text AS observation_type,
        o.status AS status,
        f.name_i18n->>'fr' AS subject,
        o.contributor_id AS contributor_id,
        u.email AS contributor_email,
        o.amount_per_litre::text AS amount,
        o.currency AS currency,
        NULL::text AS quantity,
        NULL::varchar AS unit_code,
        NULL::boolean AS is_promotion,
        o.location_id AS location_id,
        l.name AS location_name,
        g.id AS geo_area_id,
        g.name AS geo_area_name,
        s.id AS source_id,
        s.name AS source_name,
        s.kind AS source_kind,
        o.observed_at AS observed_at,
        o.confidence_score AS confidence_score,
        o.created_at AS created_at
    FROM fuel_price_observations AS o
    JOIN fuel_types AS f ON f.id = o.fuel_type_id
    JOIN locations AS l ON l.id = o.location_id
    JOIN geo_areas AS g ON g.id = l.geo_area_id
    JOIN sources AS s ON s.id = o.source_id
    LEFT JOIN users AS u ON u.id = o.contributor_id
    UNION ALL
    SELECT
        o.id AS id,
        'product_price'::text AS observation_type,
        o.status AS status,
        p.name AS subject,
        o.contributor_id AS contributor_id,
        u.email AS contributor_email,
        o.amount::text AS amount,
        o.currency AS currency,
        o.quantity::text AS quantity,
        o.unit_code AS unit_code,
        o.is_promotion AS is_promotion,
        o.location_id AS location_id,
        l.name AS location_name,
        g.id AS geo_area_id,
        g.name AS geo_area_name,
        s.id AS source_id,
        s.name AS source_name,
        s.kind AS source_kind,
        o.observed_at AS observed_at,
        o.confidence_score AS confidence_score,
        o.created_at AS created_at
    FROM product_price_observations AS o
    JOIN products AS p ON p.id = o.product_id
    JOIN locations AS l ON l.id = o.location_id
    JOIN geo_areas AS g ON g.id = l.geo_area_id
    JOIN sources AS s ON s.id = o.source_id
    LEFT JOIN users AS u ON u.id = o.contributor_id
) AS queue
WHERE (sqlc.narg('status')::text IS NULL OR queue.status::text = sqlc.narg('status')::text)
ORDER BY queue.created_at DESC, queue.id
LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;

-- name: SetProductContributionStatus :execrows
UPDATE product_price_observations
SET status = sqlc.arg('status')::moderation_status
WHERE id = sqlc.arg('id')::uuid
  AND status IN ('pending', 'flagged');

-- name: SetFuelContributionStatus :execrows
UPDATE fuel_price_observations
SET status = sqlc.arg('status')::moderation_status
WHERE id = sqlc.arg('id')::uuid
  AND status IN ('pending', 'flagged');