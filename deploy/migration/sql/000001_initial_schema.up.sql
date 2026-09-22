BEGIN;

CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('member', 'moderator', 'admin', 'partner');
CREATE TYPE moderation_status AS ENUM ('pending', 'approved', 'rejected', 'flagged');
CREATE TYPE source_kind AS ENUM ('community', 'official', 'partner');
CREATE TYPE housing_transaction AS ENUM ('rent', 'sale');
CREATE TYPE income_period AS ENUM ('monthly', 'annual');
CREATE TYPE income_basis AS ENUM ('gross', 'net');

CREATE TABLE users (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	email text NOT NULL UNIQUE,
	password_hash text,
	display_name text,
	role user_role NOT NULL DEFAULT 'member',
	locale varchar(10) NOT NULL DEFAULT 'fr',
	reputation_score integer NOT NULL DEFAULT 0,
	email_verified_at timestamptz,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	deleted_at timestamptz
);

CREATE TABLE geo_areas (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	parent_id uuid REFERENCES geo_areas(id),
	type varchar(20) NOT NULL CHECK (type IN ('country', 'region', 'city', 'postal_area')),
	code varchar(32),
	name text NOT NULL,
	country_code char(2) NOT NULL,
	geometry geometry(MultiPolygon, 4326),
	centroid geography(Point, 4326),
	UNIQUE (type, country_code, code)
);

CREATE TABLE sources (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	name text NOT NULL,
	kind source_kind NOT NULL,
	homepage_url text,
	license_name text,
	license_url text,
	attribution text,
	is_active boolean NOT NULL DEFAULT true,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE categories (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	parent_id uuid REFERENCES categories(id),
	slug text NOT NULL UNIQUE,
	name_i18n jsonb NOT NULL
);

CREATE TABLE units (
	code varchar(16) PRIMARY KEY,
	dimension varchar(16) NOT NULL CHECK (dimension IN ('mass', 'volume', 'count', 'area', 'energy')),
	to_base_factor numeric(20, 8) NOT NULL CHECK (to_base_factor > 0)
);

CREATE TABLE products (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	category_id uuid NOT NULL REFERENCES categories(id),
	name text NOT NULL,
	brand text,
	barcode varchar(32),
	reference_unit_code varchar(16) NOT NULL REFERENCES units(code),
	is_generic boolean NOT NULL DEFAULT false,
	attributes jsonb NOT NULL DEFAULT '{}',
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE merchants (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	name text NOT NULL,
	website_url text,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE locations (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	merchant_id uuid REFERENCES merchants(id),
	geo_area_id uuid NOT NULL REFERENCES geo_areas(id),
	name text NOT NULL,
	address text,
	position geography(Point, 4326),
	external_ref text,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE product_price_observations (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	product_id uuid NOT NULL REFERENCES products(id),
	location_id uuid NOT NULL REFERENCES locations(id),
	source_id uuid NOT NULL REFERENCES sources(id),
	contributor_id uuid REFERENCES users(id) ON DELETE SET NULL,
	amount numeric(12, 2) NOT NULL CHECK (amount >= 0),
	currency char(3) NOT NULL DEFAULT 'EUR',
	quantity numeric(12, 3) NOT NULL CHECK (quantity > 0),
	unit_code varchar(16) NOT NULL REFERENCES units(code),
	normalized_amount numeric(14, 4) NOT NULL CHECK (normalized_amount >= 0),
	is_promotion boolean NOT NULL DEFAULT false,
	observed_at timestamptz NOT NULL,
	status moderation_status NOT NULL DEFAULT 'pending',
	confidence_score numeric(4, 3) CHECK (confidence_score BETWEEN 0 AND 1),
	source_record_id text,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE fuel_types (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	code varchar(32) NOT NULL UNIQUE,
	name_i18n jsonb NOT NULL,
	energy varchar(16) NOT NULL
);

CREATE TABLE fuel_price_observations (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	fuel_type_id uuid NOT NULL REFERENCES fuel_types(id),
	location_id uuid NOT NULL REFERENCES locations(id),
	source_id uuid NOT NULL REFERENCES sources(id),
	contributor_id uuid REFERENCES users(id) ON DELETE SET NULL,
	amount_per_litre numeric(8, 4) NOT NULL CHECK (amount_per_litre > 0),
	currency char(3) NOT NULL DEFAULT 'EUR',
	observed_at timestamptz NOT NULL,
	status moderation_status NOT NULL DEFAULT 'pending',
	confidence_score numeric(4, 3) CHECK (confidence_score BETWEEN 0 AND 1),
	source_record_id text,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE housing_observations (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	geo_area_id uuid NOT NULL REFERENCES geo_areas(id),
	source_id uuid NOT NULL REFERENCES sources(id),
	transaction_type housing_transaction NOT NULL,
	property_type varchar(24) NOT NULL,
	rooms smallint CHECK (rooms > 0),
	furnished boolean,
	amount numeric(14, 2) NOT NULL CHECK (amount > 0),
	amount_per_sqm numeric(12, 2) CHECK (amount_per_sqm > 0),
	currency char(3) NOT NULL DEFAULT 'EUR',
	surface_sqm numeric(10, 2) CHECK (surface_sqm > 0),
	observed_at timestamptz NOT NULL,
	status moderation_status NOT NULL DEFAULT 'pending',
	source_record_id text,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE income_observations (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	geo_area_id uuid NOT NULL REFERENCES geo_areas(id),
	source_id uuid NOT NULL REFERENCES sources(id),
	occupation_code varchar(32),
	industry_code varchar(32),
	amount numeric(14, 2) NOT NULL CHECK (amount > 0),
	currency char(3) NOT NULL DEFAULT 'EUR',
	period income_period NOT NULL,
	basis income_basis NOT NULL,
	statistic varchar(16) NOT NULL CHECK (statistic IN ('mean', 'median')),
	sample_size integer CHECK (sample_size > 0),
	period_start date NOT NULL,
	period_end date NOT NULL,
	status moderation_status NOT NULL DEFAULT 'pending',
	source_record_id text,
	created_at timestamptz NOT NULL DEFAULT now(),
	CHECK (period_end >= period_start)
);

CREATE TABLE evidence_files (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	observation_type varchar(24) NOT NULL,
	observation_id uuid NOT NULL,
	object_key text NOT NULL UNIQUE,
	media_type text NOT NULL,
	sha256 char(64) NOT NULL,
	uploaded_by uuid REFERENCES users(id) ON DELETE SET NULL,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE moderation_events (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	observation_type varchar(24) NOT NULL,
	observation_id uuid NOT NULL,
	moderator_id uuid REFERENCES users(id) ON DELETE SET NULL,
	previous_status moderation_status,
	new_status moderation_status NOT NULL,
	reason_code varchar(32),
	note text,
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE price_aggregates_monthly (
	geo_area_id uuid NOT NULL REFERENCES geo_areas(id),
	metric_type varchar(24) NOT NULL,
	subject_id uuid NOT NULL,
	month date NOT NULL,
	median_amount numeric(14, 4) NOT NULL,
	min_amount numeric(14, 4) NOT NULL,
	max_amount numeric(14, 4) NOT NULL,
	sample_size integer NOT NULL CHECK (sample_size > 0),
	confidence_score numeric(4, 3) NOT NULL CHECK (confidence_score BETWEEN 0 AND 1),
	calculated_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (geo_area_id, metric_type, subject_id, month)
);

CREATE INDEX geo_areas_geometry_gix ON geo_areas USING gist (geometry);
CREATE INDEX locations_position_gix ON locations USING gist (position);
CREATE UNIQUE INDEX products_barcode_uidx
	ON products (barcode) WHERE barcode IS NOT NULL;
CREATE UNIQUE INDEX product_prices_source_record_uidx
	ON product_price_observations (source_id, source_record_id)
	WHERE source_record_id IS NOT NULL;
CREATE UNIQUE INDEX fuel_prices_source_record_uidx
	ON fuel_price_observations (source_id, source_record_id)
	WHERE source_record_id IS NOT NULL;
CREATE UNIQUE INDEX housing_source_record_uidx
	ON housing_observations (source_id, source_record_id)
	WHERE source_record_id IS NOT NULL;
CREATE UNIQUE INDEX income_source_record_uidx
	ON income_observations (source_id, source_record_id)
	WHERE source_record_id IS NOT NULL;
CREATE INDEX product_prices_lookup_idx
	ON product_price_observations (product_id, observed_at DESC)
	WHERE status = 'approved';
CREATE INDEX fuel_prices_lookup_idx
	ON fuel_price_observations (fuel_type_id, observed_at DESC)
	WHERE status = 'approved';
CREATE INDEX housing_lookup_idx
	ON housing_observations (geo_area_id, transaction_type, observed_at DESC)
	WHERE status = 'approved';
CREATE INDEX income_lookup_idx
	ON income_observations (geo_area_id, period_end DESC)
	WHERE status = 'approved';

COMMIT;
