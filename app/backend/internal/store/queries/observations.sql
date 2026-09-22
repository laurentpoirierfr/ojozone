-- name: UpsertFuelPriceBySourceRecord :one
INSERT INTO fuel_price_observations (
    fuel_type_id,
    location_id,
    source_id,
    contributor_id,
    amount_per_litre,
    currency,
    observed_at,
    status,
    confidence_score,
    source_record_id
) VALUES (
    sqlc.arg('fuel_type_id')::uuid,
    sqlc.arg('location_id')::uuid,
    sqlc.arg('source_id')::uuid,
    sqlc.narg('contributor_id')::uuid,
    sqlc.arg('amount_per_litre')::numeric,
    sqlc.arg('currency')::char(3),
    sqlc.arg('observed_at')::timestamptz,
    sqlc.arg('status')::moderation_status,
    sqlc.narg('confidence_score')::numeric,
    sqlc.arg('source_record_id')::text
)
ON CONFLICT (source_id, source_record_id) WHERE source_record_id IS NOT NULL DO UPDATE SET
    fuel_type_id = EXCLUDED.fuel_type_id,
    location_id = EXCLUDED.location_id,
    contributor_id = EXCLUDED.contributor_id,
    amount_per_litre = EXCLUDED.amount_per_litre,
    currency = EXCLUDED.currency,
    observed_at = EXCLUDED.observed_at,
    status = EXCLUDED.status,
    confidence_score = EXCLUDED.confidence_score
RETURNING *;

-- name: UpsertHousingBySourceRecord :one
INSERT INTO housing_observations (
    geo_area_id,
    source_id,
    transaction_type,
    property_type,
    rooms,
    furnished,
    amount,
    amount_per_sqm,
    currency,
    surface_sqm,
    observed_at,
    status,
    source_record_id
) VALUES (
    sqlc.arg('geo_area_id')::uuid,
    sqlc.arg('source_id')::uuid,
    sqlc.arg('transaction_type')::housing_transaction,
    sqlc.arg('property_type')::varchar,
    sqlc.narg('rooms')::smallint,
    sqlc.narg('furnished')::boolean,
    sqlc.arg('amount')::numeric,
    sqlc.narg('amount_per_sqm')::numeric,
    sqlc.arg('currency')::char(3),
    sqlc.narg('surface_sqm')::numeric,
    sqlc.arg('observed_at')::timestamptz,
    sqlc.arg('status')::moderation_status,
    sqlc.arg('source_record_id')::text
)
ON CONFLICT (source_id, source_record_id) WHERE source_record_id IS NOT NULL DO UPDATE SET
    geo_area_id = EXCLUDED.geo_area_id,
    transaction_type = EXCLUDED.transaction_type,
    property_type = EXCLUDED.property_type,
    rooms = EXCLUDED.rooms,
    furnished = EXCLUDED.furnished,
    amount = EXCLUDED.amount,
    amount_per_sqm = EXCLUDED.amount_per_sqm,
    currency = EXCLUDED.currency,
    surface_sqm = EXCLUDED.surface_sqm,
    observed_at = EXCLUDED.observed_at,
    status = EXCLUDED.status
RETURNING *;

-- name: UpsertIncomeBySourceRecord :one
INSERT INTO income_observations (
    geo_area_id,
    source_id,
    occupation_code,
    industry_code,
    amount,
    currency,
    period,
    basis,
    statistic,
    sample_size,
    period_start,
    period_end,
    status,
    source_record_id
) VALUES (
    sqlc.arg('geo_area_id')::uuid,
    sqlc.arg('source_id')::uuid,
    sqlc.narg('occupation_code')::varchar,
    sqlc.narg('industry_code')::varchar,
    sqlc.arg('amount')::numeric,
    sqlc.arg('currency')::char(3),
    sqlc.arg('period')::income_period,
    sqlc.arg('basis')::income_basis,
    sqlc.arg('statistic')::varchar,
    sqlc.narg('sample_size')::integer,
    sqlc.arg('period_start')::date,
    sqlc.arg('period_end')::date,
    sqlc.arg('status')::moderation_status,
    sqlc.arg('source_record_id')::text
)
ON CONFLICT (source_id, source_record_id) WHERE source_record_id IS NOT NULL DO UPDATE SET
    geo_area_id = EXCLUDED.geo_area_id,
    occupation_code = EXCLUDED.occupation_code,
    industry_code = EXCLUDED.industry_code,
    amount = EXCLUDED.amount,
    currency = EXCLUDED.currency,
    period = EXCLUDED.period,
    basis = EXCLUDED.basis,
    statistic = EXCLUDED.statistic,
    sample_size = EXCLUDED.sample_size,
    period_start = EXCLUDED.period_start,
    period_end = EXCLUDED.period_end,
    status = EXCLUDED.status
RETURNING *;

-- name: UpsertEvidenceByObjectKey :one
INSERT INTO evidence_files (
    observation_type, observation_id, object_key, media_type, sha256, uploaded_by
) VALUES (
    sqlc.arg('observation_type')::varchar,
    sqlc.arg('observation_id')::uuid,
    sqlc.arg('object_key')::text,
    sqlc.arg('media_type')::text,
    sqlc.arg('sha256')::char(64),
    sqlc.narg('uploaded_by')::uuid
)
ON CONFLICT (object_key) DO UPDATE SET
    observation_type = EXCLUDED.observation_type,
    observation_id = EXCLUDED.observation_id,
    media_type = EXCLUDED.media_type,
    sha256 = EXCLUDED.sha256,
    uploaded_by = EXCLUDED.uploaded_by
RETURNING *;

-- name: UpsertMonthlyPriceAggregate :one
INSERT INTO price_aggregates_monthly (
    geo_area_id,
    metric_type,
    subject_id,
    month,
    median_amount,
    min_amount,
    max_amount,
    sample_size,
    confidence_score,
    calculated_at
) VALUES (
    sqlc.arg('geo_area_id')::uuid,
    sqlc.arg('metric_type')::varchar,
    sqlc.arg('subject_id')::uuid,
    sqlc.arg('month')::date,
    sqlc.arg('median_amount')::numeric,
    sqlc.arg('min_amount')::numeric,
    sqlc.arg('max_amount')::numeric,
    sqlc.arg('sample_size')::integer,
    sqlc.arg('confidence_score')::numeric,
    sqlc.arg('calculated_at')::timestamptz
)
ON CONFLICT (geo_area_id, metric_type, subject_id, month) DO UPDATE SET
    median_amount = EXCLUDED.median_amount,
    min_amount = EXCLUDED.min_amount,
    max_amount = EXCLUDED.max_amount,
    sample_size = EXCLUDED.sample_size,
    confidence_score = EXCLUDED.confidence_score,
    calculated_at = EXCLUDED.calculated_at
RETURNING *;

-- name: ListFuelPrices :many
SELECT * FROM fuel_price_observations ORDER BY observed_at DESC,id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetFuelPrice :one
SELECT * FROM fuel_price_observations WHERE id=sqlc.arg('id')::uuid;
-- name: UpsertFuelPriceByID :one
INSERT INTO fuel_price_observations(id,fuel_type_id,location_id,source_id,contributor_id,amount_per_litre,currency,observed_at,status,confidence_score,source_record_id)
VALUES(sqlc.arg('id')::uuid,sqlc.arg('fuel_type_id')::uuid,sqlc.arg('location_id')::uuid,sqlc.arg('source_id')::uuid,sqlc.narg('contributor_id')::uuid,
sqlc.arg('amount_per_litre')::numeric,sqlc.arg('currency')::char(3),sqlc.arg('observed_at')::timestamptz,sqlc.arg('status')::moderation_status,
sqlc.narg('confidence_score')::numeric,sqlc.narg('source_record_id')::text)
ON CONFLICT(id) DO UPDATE SET fuel_type_id=EXCLUDED.fuel_type_id,location_id=EXCLUDED.location_id,source_id=EXCLUDED.source_id,
contributor_id=EXCLUDED.contributor_id,amount_per_litre=EXCLUDED.amount_per_litre,currency=EXCLUDED.currency,observed_at=EXCLUDED.observed_at,
status=EXCLUDED.status,confidence_score=EXCLUDED.confidence_score,source_record_id=EXCLUDED.source_record_id RETURNING *;
-- name: DeleteFuelPrice :execrows
DELETE FROM fuel_price_observations WHERE id=sqlc.arg('id')::uuid;

-- name: ListHousingObservations :many
SELECT * FROM housing_observations ORDER BY observed_at DESC,id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetHousingObservation :one
SELECT * FROM housing_observations WHERE id=sqlc.arg('id')::uuid;
-- name: UpsertHousingByID :one
INSERT INTO housing_observations(id,geo_area_id,source_id,transaction_type,property_type,rooms,furnished,amount,amount_per_sqm,currency,surface_sqm,observed_at,status,source_record_id)
VALUES(sqlc.arg('id')::uuid,sqlc.arg('geo_area_id')::uuid,sqlc.arg('source_id')::uuid,sqlc.arg('transaction_type')::housing_transaction,
sqlc.arg('property_type')::varchar,sqlc.narg('rooms')::smallint,sqlc.narg('furnished')::boolean,sqlc.arg('amount')::numeric,
sqlc.narg('amount_per_sqm')::numeric,sqlc.arg('currency')::char(3),sqlc.narg('surface_sqm')::numeric,sqlc.arg('observed_at')::timestamptz,
sqlc.arg('status')::moderation_status,sqlc.narg('source_record_id')::text)
ON CONFLICT(id) DO UPDATE SET geo_area_id=EXCLUDED.geo_area_id,source_id=EXCLUDED.source_id,transaction_type=EXCLUDED.transaction_type,
property_type=EXCLUDED.property_type,rooms=EXCLUDED.rooms,furnished=EXCLUDED.furnished,amount=EXCLUDED.amount,amount_per_sqm=EXCLUDED.amount_per_sqm,
currency=EXCLUDED.currency,surface_sqm=EXCLUDED.surface_sqm,observed_at=EXCLUDED.observed_at,status=EXCLUDED.status,source_record_id=EXCLUDED.source_record_id RETURNING *;
-- name: DeleteHousingObservation :execrows
DELETE FROM housing_observations WHERE id=sqlc.arg('id')::uuid;

-- name: ListIncomeObservations :many
SELECT * FROM income_observations ORDER BY period_end DESC,id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetIncomeObservation :one
SELECT * FROM income_observations WHERE id=sqlc.arg('id')::uuid;
-- name: UpsertIncomeByID :one
INSERT INTO income_observations(id,geo_area_id,source_id,occupation_code,industry_code,amount,currency,period,basis,statistic,sample_size,period_start,period_end,status,source_record_id)
VALUES(sqlc.arg('id')::uuid,sqlc.arg('geo_area_id')::uuid,sqlc.arg('source_id')::uuid,sqlc.narg('occupation_code')::varchar,
sqlc.narg('industry_code')::varchar,sqlc.arg('amount')::numeric,sqlc.arg('currency')::char(3),sqlc.arg('period')::income_period,
sqlc.arg('basis')::income_basis,sqlc.arg('statistic')::varchar,sqlc.narg('sample_size')::integer,sqlc.arg('period_start')::date,
sqlc.arg('period_end')::date,sqlc.arg('status')::moderation_status,sqlc.narg('source_record_id')::text)
ON CONFLICT(id) DO UPDATE SET geo_area_id=EXCLUDED.geo_area_id,source_id=EXCLUDED.source_id,occupation_code=EXCLUDED.occupation_code,
industry_code=EXCLUDED.industry_code,amount=EXCLUDED.amount,currency=EXCLUDED.currency,period=EXCLUDED.period,basis=EXCLUDED.basis,
statistic=EXCLUDED.statistic,sample_size=EXCLUDED.sample_size,period_start=EXCLUDED.period_start,period_end=EXCLUDED.period_end,
status=EXCLUDED.status,source_record_id=EXCLUDED.source_record_id RETURNING *;
-- name: DeleteIncomeObservation :execrows
DELETE FROM income_observations WHERE id=sqlc.arg('id')::uuid;

-- name: ListEvidenceFiles :many
SELECT * FROM evidence_files ORDER BY created_at DESC,id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetEvidenceFile :one
SELECT * FROM evidence_files WHERE id=sqlc.arg('id')::uuid;
-- name: UpsertEvidenceByID :one
INSERT INTO evidence_files(id,observation_type,observation_id,object_key,media_type,sha256,uploaded_by)
VALUES(sqlc.arg('id')::uuid,sqlc.arg('observation_type')::varchar,sqlc.arg('observation_id')::uuid,sqlc.arg('object_key')::text,
sqlc.arg('media_type')::text,sqlc.arg('sha256')::char(64),sqlc.narg('uploaded_by')::uuid)
ON CONFLICT(id) DO UPDATE SET observation_type=EXCLUDED.observation_type,observation_id=EXCLUDED.observation_id,object_key=EXCLUDED.object_key,
media_type=EXCLUDED.media_type,sha256=EXCLUDED.sha256,uploaded_by=EXCLUDED.uploaded_by RETURNING *;
-- name: DeleteEvidenceFile :execrows
DELETE FROM evidence_files WHERE id=sqlc.arg('id')::uuid;

-- name: ListPriceAggregates :many
SELECT * FROM price_aggregates_monthly ORDER BY month DESC,geo_area_id,metric_type,subject_id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetPriceAggregate :one
SELECT * FROM price_aggregates_monthly WHERE geo_area_id=sqlc.arg('geo_area_id')::uuid AND metric_type=sqlc.arg('metric_type')::varchar
AND subject_id=sqlc.arg('subject_id')::uuid AND month=sqlc.arg('month')::date;
-- name: DeletePriceAggregate :execrows
DELETE FROM price_aggregates_monthly WHERE geo_area_id=sqlc.arg('geo_area_id')::uuid AND metric_type=sqlc.arg('metric_type')::varchar
AND subject_id=sqlc.arg('subject_id')::uuid AND month=sqlc.arg('month')::date;

-- name: ListModerationEvents :many
SELECT * FROM moderation_events ORDER BY created_at DESC,id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetModerationEvent :one
SELECT * FROM moderation_events WHERE id=sqlc.arg('id')::uuid;
-- name: AppendModerationEvent :one
INSERT INTO moderation_events(observation_type,observation_id,moderator_id,previous_status,new_status,reason_code,note)
VALUES(sqlc.arg('observation_type')::varchar,sqlc.arg('observation_id')::uuid,sqlc.narg('moderator_id')::uuid,
sqlc.narg('previous_status')::moderation_status,sqlc.arg('new_status')::moderation_status,sqlc.narg('reason_code')::varchar,sqlc.narg('note')::text) RETURNING *;

-- name: ListAdminUsers :many
SELECT id,email,display_name,role,locale,reputation_score,email_verified_at,created_at,updated_at,deleted_at
FROM users ORDER BY created_at DESC,id LIMIT sqlc.arg('page_size')::integer OFFSET sqlc.arg('page_offset')::integer;
-- name: GetAdminUser :one
SELECT id,email,display_name,role,locale,reputation_score,email_verified_at,created_at,updated_at,deleted_at FROM users WHERE id=sqlc.arg('id')::uuid;
