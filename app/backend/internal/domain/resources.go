package domain

import (
	"encoding/json"
	"time"
)

type Pagination struct{ Limit, Offset int32 }
type ResourceKey struct{ ID, Code, GeoAreaID, MetricType, SubjectID, Month string }

type GeoArea struct {
	ID          string   `json:"id"`
	ParentID    *string  `json:"parent_id"`
	Type        string   `json:"type"`
	Code        *string  `json:"code"`
	Name        string   `json:"name"`
	CountryCode string   `json:"country_code"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}
type GeoAreaUpsert struct {
	ID          string   `json:"id,omitempty"`
	ParentID    *string  `json:"parent_id,omitempty"`
	Type        string   `json:"type"`
	Code        *string  `json:"code,omitempty"`
	Name        string   `json:"name"`
	CountryCode string   `json:"country_code"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

type Source struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Kind        string    `json:"kind"`
	HomepageURL *string   `json:"homepage_url"`
	LicenseName *string   `json:"license_name"`
	LicenseURL  *string   `json:"license_url"`
	Attribution *string   `json:"attribution"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
type SourceUpsert struct {
	ID          string  `json:"id,omitempty"`
	Name        string  `json:"name"`
	Kind        string  `json:"kind"`
	HomepageURL *string `json:"homepage_url,omitempty"`
	LicenseName *string `json:"license_name,omitempty"`
	LicenseURL  *string `json:"license_url,omitempty"`
	Attribution *string `json:"attribution,omitempty"`
	IsActive    bool    `json:"is_active"`
}

type Category struct {
	ID       string          `json:"id"`
	ParentID *string         `json:"parent_id"`
	Slug     string          `json:"slug"`
	NameI18n json.RawMessage `json:"name_i18n" swaggertype:"object"`
}
type CategoryUpsert struct {
	ID       string          `json:"id,omitempty"`
	ParentID *string         `json:"parent_id,omitempty"`
	Slug     string          `json:"slug"`
	NameI18n json.RawMessage `json:"name_i18n" swaggertype:"object"`
}
type Unit struct {
	Code         string `json:"code"`
	Dimension    string `json:"dimension"`
	ToBaseFactor string `json:"to_base_factor"`
}
type UnitUpsert = Unit
type Merchant struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	WebsiteURL *string   `json:"website_url"`
	CreatedAt  time.Time `json:"created_at"`
}
type MerchantUpsert struct {
	ID         string  `json:"id,omitempty"`
	Name       string  `json:"name"`
	WebsiteURL *string `json:"website_url,omitempty"`
}
type Location struct {
	ID          string    `json:"id"`
	MerchantID  *string   `json:"merchant_id"`
	GeoAreaID   string    `json:"geo_area_id"`
	Name        string    `json:"name"`
	Address     *string   `json:"address"`
	Latitude    *float64  `json:"latitude"`
	Longitude   *float64  `json:"longitude"`
	ExternalRef *string   `json:"external_ref"`
	CreatedAt   time.Time `json:"created_at"`
}
type LocationUpsert struct {
	ID          string   `json:"id,omitempty"`
	MerchantID  *string  `json:"merchant_id,omitempty"`
	GeoAreaID   string   `json:"geo_area_id"`
	Name        string   `json:"name"`
	Address     *string  `json:"address,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	ExternalRef *string  `json:"external_ref,omitempty"`
}
type FuelType struct {
	ID       string          `json:"id"`
	Code     string          `json:"code"`
	NameI18n json.RawMessage `json:"name_i18n" swaggertype:"object"`
	Energy   string          `json:"energy"`
}
type FuelTypeUpsert struct {
	ID       string          `json:"id,omitempty"`
	Code     string          `json:"code"`
	NameI18n json.RawMessage `json:"name_i18n" swaggertype:"object"`
	Energy   string          `json:"energy"`
}

type FuelPrice struct {
	ID              string    `json:"id"`
	FuelTypeID      string    `json:"fuel_type_id"`
	LocationID      string    `json:"location_id"`
	SourceID        string    `json:"source_id"`
	ContributorID   *string   `json:"contributor_id"`
	AmountPerLitre  string    `json:"amount_per_litre"`
	Currency        string    `json:"currency"`
	ObservedAt      time.Time `json:"observed_at"`
	Status          string    `json:"status"`
	ConfidenceScore *string   `json:"confidence_score"`
	SourceRecordID  *string   `json:"source_record_id"`
	CreatedAt       time.Time `json:"created_at"`
}
type FuelPriceUpsert struct {
	ID              string    `json:"id,omitempty"`
	FuelTypeID      string    `json:"fuel_type_id"`
	LocationID      string    `json:"location_id"`
	SourceID        string    `json:"source_id"`
	ContributorID   *string   `json:"contributor_id,omitempty"`
	AmountPerLitre  string    `json:"amount_per_litre"`
	Currency        string    `json:"currency"`
	ObservedAt      time.Time `json:"observed_at"`
	Status          string    `json:"status"`
	ConfidenceScore *string   `json:"confidence_score,omitempty"`
	SourceRecordID  *string   `json:"source_record_id,omitempty"`
}
type HousingObservation struct {
	ID              string    `json:"id"`
	GeoAreaID       string    `json:"geo_area_id"`
	SourceID        string    `json:"source_id"`
	TransactionType string    `json:"transaction_type"`
	PropertyType    string    `json:"property_type"`
	Rooms           *int16    `json:"rooms"`
	Furnished       *bool     `json:"furnished"`
	Amount          string    `json:"amount"`
	AmountPerSqm    *string   `json:"amount_per_sqm"`
	Currency        string    `json:"currency"`
	SurfaceSqm      *string   `json:"surface_sqm"`
	ObservedAt      time.Time `json:"observed_at"`
	Status          string    `json:"status"`
	SourceRecordID  *string   `json:"source_record_id"`
	CreatedAt       time.Time `json:"created_at"`
}
type HousingUpsert struct {
	ID              string    `json:"id,omitempty"`
	GeoAreaID       string    `json:"geo_area_id"`
	SourceID        string    `json:"source_id"`
	TransactionType string    `json:"transaction_type"`
	PropertyType    string    `json:"property_type"`
	Rooms           *int16    `json:"rooms,omitempty"`
	Furnished       *bool     `json:"furnished,omitempty"`
	Amount          string    `json:"amount"`
	AmountPerSqm    *string   `json:"amount_per_sqm,omitempty"`
	Currency        string    `json:"currency"`
	SurfaceSqm      *string   `json:"surface_sqm,omitempty"`
	ObservedAt      time.Time `json:"observed_at"`
	Status          string    `json:"status"`
	SourceRecordID  *string   `json:"source_record_id,omitempty"`
}
type IncomeObservation struct {
	ID             string    `json:"id"`
	GeoAreaID      string    `json:"geo_area_id"`
	SourceID       string    `json:"source_id"`
	OccupationCode *string   `json:"occupation_code"`
	IndustryCode   *string   `json:"industry_code"`
	Amount         string    `json:"amount"`
	Currency       string    `json:"currency"`
	Period         string    `json:"period"`
	Basis          string    `json:"basis"`
	Statistic      string    `json:"statistic"`
	SampleSize     *int32    `json:"sample_size"`
	PeriodStart    string    `json:"period_start"`
	PeriodEnd      string    `json:"period_end"`
	Status         string    `json:"status"`
	SourceRecordID *string   `json:"source_record_id"`
	CreatedAt      time.Time `json:"created_at"`
}
type IncomeUpsert struct {
	ID             string  `json:"id,omitempty"`
	GeoAreaID      string  `json:"geo_area_id"`
	SourceID       string  `json:"source_id"`
	OccupationCode *string `json:"occupation_code,omitempty"`
	IndustryCode   *string `json:"industry_code,omitempty"`
	Amount         string  `json:"amount"`
	Currency       string  `json:"currency"`
	Period         string  `json:"period"`
	Basis          string  `json:"basis"`
	Statistic      string  `json:"statistic"`
	SampleSize     *int32  `json:"sample_size,omitempty"`
	PeriodStart    string  `json:"period_start"`
	PeriodEnd      string  `json:"period_end"`
	Status         string  `json:"status"`
	SourceRecordID *string `json:"source_record_id,omitempty"`
}
type EvidenceFile struct {
	ID              string    `json:"id"`
	ObservationType string    `json:"observation_type"`
	ObservationID   string    `json:"observation_id"`
	ObjectKey       string    `json:"object_key"`
	MediaType       string    `json:"media_type"`
	SHA256          string    `json:"sha256"`
	UploadedBy      *string   `json:"uploaded_by"`
	CreatedAt       time.Time `json:"created_at"`
}
type EvidenceUpsert struct {
	ID              string  `json:"id,omitempty"`
	ObservationType string  `json:"observation_type"`
	ObservationID   string  `json:"observation_id"`
	ObjectKey       string  `json:"object_key"`
	MediaType       string  `json:"media_type"`
	SHA256          string  `json:"sha256"`
	UploadedBy      *string `json:"uploaded_by,omitempty"`
}
type PriceAggregate struct {
	GeoAreaID       string    `json:"geo_area_id"`
	MetricType      string    `json:"metric_type"`
	SubjectID       string    `json:"subject_id"`
	Month           string    `json:"month"`
	MedianAmount    string    `json:"median_amount"`
	MinAmount       string    `json:"min_amount"`
	MaxAmount       string    `json:"max_amount"`
	SampleSize      int32     `json:"sample_size"`
	ConfidenceScore string    `json:"confidence_score"`
	CalculatedAt    time.Time `json:"calculated_at"`
}
type PriceAggregateUpsert = PriceAggregate
type ModerationEvent struct {
	ID              string    `json:"id"`
	ObservationType string    `json:"observation_type"`
	ObservationID   string    `json:"observation_id"`
	ModeratorID     *string   `json:"moderator_id"`
	PreviousStatus  *string   `json:"previous_status"`
	NewStatus       string    `json:"new_status"`
	ReasonCode      *string   `json:"reason_code"`
	Note            *string   `json:"note"`
	CreatedAt       time.Time `json:"created_at"`
}
type ModerationEventAppend struct {
	ObservationType string  `json:"observation_type"`
	ObservationID   string  `json:"observation_id"`
	ModeratorID     *string `json:"moderator_id,omitempty"`
	PreviousStatus  *string `json:"previous_status,omitempty"`
	NewStatus       string  `json:"new_status"`
	ReasonCode      *string `json:"reason_code,omitempty"`
	Note            *string `json:"note,omitempty"`
}
type AdminUser struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	DisplayName     *string    `json:"display_name"`
	Role            string     `json:"role"`
	Locale          string     `json:"locale"`
	ReputationScore int32      `json:"reputation_score"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}
