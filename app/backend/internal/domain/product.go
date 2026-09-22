package domain

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource conflict")
)

type Product struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Brand            *string         `json:"brand"`
	Barcode          *string         `json:"barcode"`
	ReferenceUnit    string          `json:"reference_unit"`
	IsGeneric        bool            `json:"is_generic"`
	Attributes       json.RawMessage `json:"attributes" swaggertype:"object"`
	CategorySlug     string          `json:"category_slug"`
	CategoryNameI18n json.RawMessage `json:"category_name_i18n" swaggertype:"object"`
	CreatedAt        time.Time       `json:"created_at"`
}

type ProductUpsert struct {
	ID            string          `json:"id,omitempty" format:"uuid"`
	CategoryID    string          `json:"category_id" format:"uuid"`
	Name          string          `json:"name"`
	Brand         *string         `json:"brand,omitempty"`
	Barcode       *string         `json:"barcode,omitempty" maxLength:"32"`
	ReferenceUnit string          `json:"reference_unit" maxLength:"16"`
	IsGeneric     bool            `json:"is_generic"`
	Attributes    json.RawMessage `json:"attributes" swaggertype:"object"`
}

type ProductPrice struct {
	ID               string    `json:"id"`
	ProductID        string    `json:"product_id"`
	Amount           string    `json:"amount"`
	Currency         string    `json:"currency"`
	Quantity         string    `json:"quantity"`
	UnitCode         string    `json:"unit_code"`
	NormalizedAmount string    `json:"normalized_amount"`
	IsPromotion      bool      `json:"is_promotion"`
	ObservedAt       time.Time `json:"observed_at"`
	Status           string    `json:"status"`
	ConfidenceScore  *string   `json:"confidence_score"`
	SourceRecordID   *string   `json:"source_record_id"`
	Location         EntityRef `json:"location"`
	GeoArea          EntityRef `json:"geo_area"`
	Source           SourceRef `json:"source"`
}

type ProductPriceUpsert struct {
	ID               string    `json:"id,omitempty" format:"uuid"`
	ProductID        string    `json:"product_id" format:"uuid"`
	LocationID       string    `json:"location_id" format:"uuid"`
	SourceID         string    `json:"source_id" format:"uuid"`
	ContributorID    *string   `json:"contributor_id,omitempty" format:"uuid"`
	Amount           string    `json:"amount" example:"2.35"`
	Currency         string    `json:"currency" example:"EUR" minLength:"3" maxLength:"3"`
	Quantity         string    `json:"quantity" example:"1.000"`
	UnitCode         string    `json:"unit_code" example:"L" maxLength:"16"`
	NormalizedAmount string    `json:"normalized_amount" example:"2.3500"`
	IsPromotion      bool      `json:"is_promotion"`
	ObservedAt       time.Time `json:"observed_at" format:"date-time"`
	Status           string    `json:"status" enums:"pending,approved,rejected,flagged"`
	ConfidenceScore  *string   `json:"confidence_score,omitempty" example:"0.850"`
	SourceRecordID   *string   `json:"source_record_id,omitempty"`
}

type EntityRef struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Address *string `json:"address,omitempty"`
}

type SourceRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type ProductFilter struct {
	Search string
	Limit  int32
	Offset int32
}

type PriceFilter struct {
	ProductID string
	GeoAreaID string
	Status    string
	Limit     int32
	Offset    int32
}
