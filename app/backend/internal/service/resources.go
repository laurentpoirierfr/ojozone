package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
)

const (
	ResourceGeoAreas   = "geo-areas"
	ResourceSources    = "sources"
	ResourceCategories = "categories"
	ResourceUnits      = "units"
	ResourceMerchants  = "merchants"
	ResourceLocations  = "locations"
	ResourceFuelTypes  = "fuel-types"
	ResourceFuelPrices = "fuel-prices"
	ResourceHousing    = "housing-observations"
	ResourceIncome     = "income-observations"
	ResourceEvidence   = "evidence-files"
	ResourceAggregates = "price-aggregates"
	ResourceModeration = "moderation-events"
	ResourceAdminUsers = "admin-users"
	ResourceProducts   = "products"
	ResourcePrices     = "product-prices"
)

func (s *OjoZone) ListResource(ctx context.Context, resource string, page domain.Pagination) (any, error) {
	if err := validatePagination(page.Limit, page.Offset); err != nil {
		return nil, err
	}
	return s.repository.ListResource(ctx, resource, page)
}

func (s *OjoZone) GetResource(ctx context.Context, resource string, key domain.ResourceKey) (any, error) {
	if err := validateResourceKey(resource, key); err != nil {
		return nil, err
	}
	return s.repository.GetResource(ctx, resource, key)
}

func (s *OjoZone) UpsertResource(ctx context.Context, resource string, raw json.RawMessage) (any, error) {
	input, err := decodeAndValidateResource(resource, raw, false, domain.ResourceKey{})
	if err != nil {
		return nil, err
	}
	return s.repository.UpsertResource(ctx, resource, input, false)
}

func (s *OjoZone) ReplaceResource(ctx context.Context, resource string, key domain.ResourceKey, raw json.RawMessage) (any, error) {
	if err := validateResourceKey(resource, key); err != nil {
		return nil, err
	}
	input, err := decodeAndValidateResource(resource, raw, true, key)
	if err != nil {
		return nil, err
	}
	return s.repository.UpsertResource(ctx, resource, input, true)
}

func (s *OjoZone) DeleteResource(ctx context.Context, resource string, key domain.ResourceKey) error {
	if err := validateResourceKey(resource, key); err != nil {
		return err
	}
	return s.repository.DeleteResource(ctx, resource, key)
}

func validateResourceKey(resource string, key domain.ResourceKey) error {
	if resource == ResourceUnits {
		if strings.TrimSpace(key.Code) == "" || len(key.Code) > 16 {
			return ErrInvalidResource
		}
		return nil
	}
	if resource == ResourceAggregates {
		if !validUUID(key.GeoAreaID) || !validUUID(key.SubjectID) || strings.TrimSpace(key.MetricType) == "" {
			return ErrInvalidResource
		}
		_, err := time.Parse("2006-01-02", key.Month)
		if err != nil {
			return ErrInvalidResource
		}
		return nil
	}
	if !validUUID(key.ID) {
		return ErrInvalidID
	}
	return nil
}

func decodeAndValidateResource(resource string, raw json.RawMessage, replace bool, key domain.ResourceKey) (any, error) {
	switch resource {
	case ResourceGeoAreas:
		var value domain.GeoAreaUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		value.Type = strings.TrimSpace(value.Type)
		value.Name = strings.TrimSpace(value.Name)
		value.CountryCode = strings.ToUpper(strings.TrimSpace(value.CountryCode))
		if value.ParentID != nil && !validUUID(*value.ParentID) || value.Name == "" || len(value.CountryCode) != 2 || !oneOf(value.Type, "country", "region", "city", "postal_area") || value.Latitude != nil != (value.Longitude != nil) || value.Latitude != nil && (*value.Latitude < -90 || *value.Latitude > 90 || *value.Longitude < -180 || *value.Longitude > 180) {
			return nil, ErrInvalidResource
		}
		if !replace && (value.Code == nil || strings.TrimSpace(*value.Code) == "") {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceSources:
		var value domain.SourceUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		value.Name = strings.TrimSpace(value.Name)
		if !validUUID(value.ID) || value.Name == "" || !oneOf(value.Kind, "community", "official", "partner") {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceCategories:
		var value domain.CategoryUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		value.Slug = strings.TrimSpace(value.Slug)
		if value.ParentID != nil && !validUUID(*value.ParentID) || value.Slug == "" || !validJSONObject(value.NameI18n) {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceUnits:
		var value domain.Unit
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.Code = key.Code
		}
		value.Code = strings.TrimSpace(value.Code)
		if value.Code == "" || len(value.Code) > 16 || !oneOf(value.Dimension, "mass", "volume", "count", "area", "energy") || !validPositiveDecimal(value.ToBaseFactor) {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceMerchants:
		var value domain.MerchantUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		value.Name = strings.TrimSpace(value.Name)
		if !validUUID(value.ID) || value.Name == "" {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceLocations:
		var value domain.LocationUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		value.Name = strings.TrimSpace(value.Name)
		if !validUUID(value.ID) || !validUUID(value.GeoAreaID) || value.MerchantID != nil && !validUUID(*value.MerchantID) || value.Name == "" || value.Latitude != nil != (value.Longitude != nil) || value.Latitude != nil && (*value.Latitude < -90 || *value.Latitude > 90 || *value.Longitude < -180 || *value.Longitude > 180) {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceFuelTypes:
		var value domain.FuelTypeUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		value.Code = strings.TrimSpace(value.Code)
		value.Energy = strings.TrimSpace(value.Energy)
		if value.Code == "" || value.Energy == "" || !validJSONObject(value.NameI18n) {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceFuelPrices:
		var value domain.FuelPriceUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		if !validUUIDs(value.FuelTypeID, value.LocationID, value.SourceID) || value.ContributorID != nil && !validUUID(*value.ContributorID) || !validPositiveDecimal(value.AmountPerLitre) || !validObservation(value.Currency, value.ObservedAt, value.Status, value.ConfidenceScore) || !replace && (value.SourceRecordID == nil || strings.TrimSpace(*value.SourceRecordID) == "") {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceHousing:
		var value domain.HousingUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		if !validUUIDs(value.GeoAreaID, value.SourceID) || !oneOf(value.TransactionType, "rent", "sale") || strings.TrimSpace(value.PropertyType) == "" || value.Rooms != nil && *value.Rooms <= 0 || !validPositiveDecimal(value.Amount) || value.AmountPerSqm != nil && !validPositiveDecimal(*value.AmountPerSqm) || value.SurfaceSqm != nil && !validPositiveDecimal(*value.SurfaceSqm) || !validObservation(value.Currency, value.ObservedAt, value.Status, nil) || !replace && (value.SourceRecordID == nil || strings.TrimSpace(*value.SourceRecordID) == "") {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceIncome:
		var value domain.IncomeUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		start, e1 := time.Parse("2006-01-02", value.PeriodStart)
		end, e2 := time.Parse("2006-01-02", value.PeriodEnd)
		if !validUUIDs(value.GeoAreaID, value.SourceID) || !validPositiveDecimal(value.Amount) || !oneOf(value.Period, "monthly", "annual") || !oneOf(value.Basis, "gross", "net") || !oneOf(value.Statistic, "mean", "median") || value.SampleSize != nil && *value.SampleSize <= 0 || e1 != nil || e2 != nil || end.Before(start) || !validObservation(value.Currency, time.Now(), value.Status, nil) || !replace && (value.SourceRecordID == nil || strings.TrimSpace(*value.SourceRecordID) == "") {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceEvidence:
		var value domain.EvidenceUpsert
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.ID = key.ID
		}
		if !validUUID(value.ObservationID) || value.UploadedBy != nil && !validUUID(*value.UploadedBy) || strings.TrimSpace(value.ObservationType) == "" || strings.TrimSpace(value.ObjectKey) == "" || strings.TrimSpace(value.MediaType) == "" || len(value.SHA256) != 64 {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceAggregates:
		var value domain.PriceAggregate
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if replace {
			value.GeoAreaID = key.GeoAreaID
			value.MetricType = key.MetricType
			value.SubjectID = key.SubjectID
			value.Month = key.Month
		}
		_, dateErr := time.Parse("2006-01-02", value.Month)
		if !validUUIDs(value.GeoAreaID, value.SubjectID) || strings.TrimSpace(value.MetricType) == "" || dateErr != nil || !validNonNegativeDecimal(value.MedianAmount) || !validNonNegativeDecimal(value.MinAmount) || !validNonNegativeDecimal(value.MaxAmount) || value.SampleSize <= 0 || !validConfidence(value.ConfidenceScore) || value.CalculatedAt.IsZero() {
			return nil, ErrInvalidResource
		}
		return value, nil
	case ResourceModeration:
		if replace {
			return nil, ErrInvalidResource
		}
		var value domain.ModerationEventAppend
		if json.Unmarshal(raw, &value) != nil {
			return nil, ErrInvalidResource
		}
		if !validUUID(value.ObservationID) || value.ModeratorID != nil && !validUUID(*value.ModeratorID) || value.PreviousStatus != nil && !validStatus(*value.PreviousStatus) || !validStatus(value.NewStatus) || strings.TrimSpace(value.ObservationType) == "" {
			return nil, ErrInvalidResource
		}
		return value, nil
	default:
		return nil, ErrInvalidResource
	}
}

func validJSONObject(value json.RawMessage) bool {
	return len(value) > 1 && json.Valid(value) && value[0] == '{'
}
func oneOf(value string, values ...string) bool {
	for _, candidate := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
func validUUIDs(values ...string) bool {
	for _, value := range values {
		if !validUUID(value) {
			return false
		}
	}
	return true
}
func validStatus(value string) bool {
	return oneOf(value, "pending", "approved", "rejected", "flagged")
}
func validObservation(currency string, observedAt time.Time, status string, confidence *string) bool {
	return len(strings.TrimSpace(currency)) == 3 && !observedAt.IsZero() && validStatus(status) && (confidence == nil || validConfidence(*confidence))
}
