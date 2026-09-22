-- name: CreateUser :one
INSERT INTO users (email, password_hash, display_name, locale)
VALUES (
    sqlc.arg('email')::text,
    sqlc.arg('password_hash')::text,
    sqlc.narg('display_name')::text,
    sqlc.arg('locale')::varchar
)
RETURNING id, email, display_name, role, locale, reputation_score, email_verified_at, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, display_name, role, locale, reputation_score, email_verified_at, created_at, updated_at, deleted_at
FROM users
WHERE email = sqlc.arg('email')::text;

-- name: GetUserById :one
SELECT id, email, display_name, role, locale, reputation_score, email_verified_at, created_at, updated_at
FROM users
WHERE id = sqlc.arg('id')::uuid;

-- name: UpdateUserProfile :one
UPDATE users
SET display_name = sqlc.narg('display_name')::text,
    locale = sqlc.arg('locale')::varchar,
    updated_at = now()
WHERE id = sqlc.arg('id')::uuid
RETURNING id, email, display_name, role, locale, reputation_score, email_verified_at, created_at, updated_at;

-- name: AnonymizeUser :one
UPDATE users
SET email = 'anon-' || gen_random_uuid()::text || '@deleted.local',
    display_name = NULL,
    locale = 'fr',
    reputation_score = 0,
    email_verified_at = NULL,
    deleted_at = now(),
    updated_at = now()
WHERE id = sqlc.arg('id')::uuid AND deleted_at IS NULL
RETURNING id, email, display_name, role, locale, reputation_score, email_verified_at, created_at, updated_at, deleted_at;

-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token_hash, expires_at, user_agent, ip_address)
VALUES (
    sqlc.arg('user_id')::uuid,
    sqlc.arg('refresh_token_hash')::char(64),
    sqlc.arg('expires_at')::timestamptz,
    sqlc.narg('user_agent')::text,
    sqlc.narg('ip_address')::inet
)
RETURNING *;

-- name: GetSessionByRefreshHash :one
SELECT * FROM sessions
WHERE refresh_token_hash = sqlc.arg('refresh_token_hash')::char(64);

-- name: UpdateSessionRefresh :one
UPDATE sessions
SET refresh_token_hash = sqlc.arg('refresh_token_hash')::char(64),
    expires_at = sqlc.arg('expires_at')::timestamptz,
    updated_at = now()
WHERE id = sqlc.arg('id')::uuid AND revoked_at IS NULL
RETURNING *;

-- name: RevokeSession :execrows
UPDATE sessions
SET revoked_at = now(), updated_at = now()
WHERE id = sqlc.arg('id')::uuid AND revoked_at IS NULL;

-- name: RevokeAllSessionsForUser :execrows
UPDATE sessions
SET revoked_at = now(), updated_at = now()
WHERE user_id = sqlc.arg('user_id')::uuid AND revoked_at IS NULL;

-- name: ListProductPriceContributions :many
SELECT
    o.id,
    o.amount,
    o.currency,
    o.quantity,
    o.unit_code,
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
WHERE o.contributor_id = sqlc.arg('contributor_id')::uuid
ORDER BY o.created_at DESC, o.id
LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;

-- name: ListFuelPriceContributions :many
SELECT
    o.id,
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
WHERE o.contributor_id = sqlc.arg('contributor_id')::uuid
ORDER BY o.created_at DESC, o.id
LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;