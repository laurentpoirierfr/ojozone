package repository

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/store/db"
)

func (r *PostgreSQL) ListResource(ctx context.Context, resource string, page domain.Pagination) (any, error) {
	switch resource {
	case "geo-areas":
		rows, err := r.queries.ListGeoAreas(ctx, db.ListGeoAreasParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.GeoArea, 0, len(rows))
		for _, row := range rows {
			out = append(out, geoArea(row.ID, row.ParentID, row.Type, row.Code, row.Name, row.CountryCode, row.Latitude, row.Longitude))
		}
		return out, nil
	case "sources":
		rows, err := r.queries.ListSources(ctx, db.ListSourcesParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.Source, 0, len(rows))
		for _, row := range rows {
			out = append(out, source(row))
		}
		return out, nil
	case "categories":
		rows, err := r.queries.ListCategories(ctx, db.ListCategoriesParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.Category, 0, len(rows))
		for _, row := range rows {
			out = append(out, category(row))
		}
		return out, nil
	case "units":
		rows, err := r.queries.ListUnits(ctx, db.ListUnitsParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.Unit, 0, len(rows))
		for _, row := range rows {
			out = append(out, unit(row))
		}
		return out, nil
	case "merchants":
		rows, err := r.queries.ListMerchants(ctx, db.ListMerchantsParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.Merchant, 0, len(rows))
		for _, row := range rows {
			out = append(out, merchant(row))
		}
		return out, nil
	case "locations":
		rows, err := r.queries.ListLocations(ctx, db.ListLocationsParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.Location, 0, len(rows))
		for _, row := range rows {
			out = append(out, location(row.ID, row.MerchantID, row.GeoAreaID, row.Name, row.Address, row.ExternalRef, row.CreatedAt, row.Latitude, row.Longitude))
		}
		return out, nil
	case "fuel-types":
		rows, err := r.queries.ListFuelTypes(ctx, db.ListFuelTypesParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.FuelType, 0, len(rows))
		for _, row := range rows {
			out = append(out, fuelType(row))
		}
		return out, nil
	case "fuel-prices":
		rows, err := r.queries.ListFuelPrices(ctx, db.ListFuelPricesParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.FuelPrice, 0, len(rows))
		for _, row := range rows {
			out = append(out, fuelPrice(row))
		}
		return out, nil
	case "housing-observations":
		rows, err := r.queries.ListHousingObservations(ctx, db.ListHousingObservationsParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.HousingObservation, 0, len(rows))
		for _, row := range rows {
			out = append(out, housing(row))
		}
		return out, nil
	case "income-observations":
		rows, err := r.queries.ListIncomeObservations(ctx, db.ListIncomeObservationsParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.IncomeObservation, 0, len(rows))
		for _, row := range rows {
			out = append(out, income(row))
		}
		return out, nil
	case "evidence-files":
		rows, err := r.queries.ListEvidenceFiles(ctx, db.ListEvidenceFilesParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.EvidenceFile, 0, len(rows))
		for _, row := range rows {
			out = append(out, evidence(row))
		}
		return out, nil
	case "price-aggregates":
		rows, err := r.queries.ListPriceAggregates(ctx, db.ListPriceAggregatesParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.PriceAggregate, 0, len(rows))
		for _, row := range rows {
			out = append(out, aggregate(row))
		}
		return out, nil
	case "moderation-events":
		rows, err := r.queries.ListModerationEvents(ctx, db.ListModerationEventsParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.ModerationEvent, 0, len(rows))
		for _, row := range rows {
			out = append(out, moderation(row))
		}
		return out, nil
	case "admin-users":
		rows, err := r.queries.ListAdminUsers(ctx, db.ListAdminUsersParams{PageSize: page.Limit, PageOffset: page.Offset})
		if err != nil {
			return nil, err
		}
		out := make([]domain.AdminUser, 0, len(rows))
		for _, row := range rows {
			out = append(out, adminUser(row.ID, row.Email, row.DisplayName, row.Role, row.Locale, row.ReputationScore, row.EmailVerifiedAt, row.CreatedAt, row.UpdatedAt, row.DeletedAt))
		}
		return out, nil
	default:
		return nil, domain.ErrNotFound
	}
}

func (r *PostgreSQL) GetResource(ctx context.Context, resource string, key domain.ResourceKey) (any, error) {
	var value any
	var err error
	switch resource {
	case "geo-areas":
		var row db.GetGeoAreaRow
		row, err = r.queries.GetGeoArea(ctx, postgresUUID(key.ID))
		value = geoArea(row.ID, row.ParentID, row.Type, row.Code, row.Name, row.CountryCode, row.Latitude, row.Longitude)
	case "sources":
		var row db.Source
		row, err = r.queries.GetSource(ctx, postgresUUID(key.ID))
		value = source(row)
	case "categories":
		var row db.Category
		row, err = r.queries.GetCategory(ctx, postgresUUID(key.ID))
		value = category(row)
	case "units":
		var row db.Unit
		row, err = r.queries.GetUnit(ctx, key.Code)
		value = unit(row)
	case "merchants":
		var row db.Merchant
		row, err = r.queries.GetMerchant(ctx, postgresUUID(key.ID))
		value = merchant(row)
	case "locations":
		var row db.GetLocationRow
		row, err = r.queries.GetLocation(ctx, postgresUUID(key.ID))
		value = location(row.ID, row.MerchantID, row.GeoAreaID, row.Name, row.Address, row.ExternalRef, row.CreatedAt, row.Latitude, row.Longitude)
	case "fuel-types":
		var row db.FuelType
		row, err = r.queries.GetFuelType(ctx, postgresUUID(key.ID))
		value = fuelType(row)
	case "fuel-prices":
		var row db.FuelPriceObservation
		row, err = r.queries.GetFuelPrice(ctx, postgresUUID(key.ID))
		value = fuelPrice(row)
	case "housing-observations":
		var row db.HousingObservation
		row, err = r.queries.GetHousingObservation(ctx, postgresUUID(key.ID))
		value = housing(row)
	case "income-observations":
		var row db.IncomeObservation
		row, err = r.queries.GetIncomeObservation(ctx, postgresUUID(key.ID))
		value = income(row)
	case "evidence-files":
		var row db.EvidenceFile
		row, err = r.queries.GetEvidenceFile(ctx, postgresUUID(key.ID))
		value = evidence(row)
	case "price-aggregates":
		var row db.PriceAggregatesMonthly
		row, err = r.queries.GetPriceAggregate(ctx, db.GetPriceAggregateParams{GeoAreaID: postgresUUID(key.GeoAreaID), MetricType: key.MetricType, SubjectID: postgresUUID(key.SubjectID), Month: postgresDate(key.Month)})
		value = aggregate(row)
	case "moderation-events":
		var row db.ModerationEvent
		row, err = r.queries.GetModerationEvent(ctx, postgresUUID(key.ID))
		value = moderation(row)
	case "admin-users":
		var row db.GetAdminUserRow
		row, err = r.queries.GetAdminUser(ctx, postgresUUID(key.ID))
		value = adminUser(row.ID, row.Email, row.DisplayName, row.Role, row.Locale, row.ReputationScore, row.EmailVerifiedAt, row.CreatedAt, row.UpdatedAt, row.DeletedAt)
	default:
		return nil, domain.ErrNotFound
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return value, err
}

func (r *PostgreSQL) UpsertResource(ctx context.Context, resource string, input any, byID bool) (any, error) {
	var value any
	var err error
	switch resource {
	case "geo-areas":
		v := input.(domain.GeoAreaUpsert)
		if byID {
			var row db.GeoArea
			row, err = r.queries.UpsertGeoAreaByID(ctx, db.UpsertGeoAreaByIDParams{ID: postgresUUID(v.ID), ParentID: nullablePostgresUUID(v.ParentID), Type: v.Type, Code: v.Code, Name: v.Name, CountryCode: v.CountryCode, Latitude: v.Latitude, Longitude: v.Longitude})
			if err == nil {
				value, err = r.GetResource(ctx, resource, domain.ResourceKey{ID: uuidString(row.ID)})
			}
		} else {
			var row db.GeoArea
			row, err = r.queries.UpsertGeoAreaByCode(ctx, db.UpsertGeoAreaByCodeParams{ParentID: nullablePostgresUUID(v.ParentID), Type: v.Type, Code: *v.Code, Name: v.Name, CountryCode: v.CountryCode, Latitude: v.Latitude, Longitude: v.Longitude})
			if err == nil {
				value, err = r.GetResource(ctx, resource, domain.ResourceKey{ID: uuidString(row.ID)})
			}
		}
	case "sources":
		v := input.(domain.SourceUpsert)
		var row db.Source
		row, err = r.queries.UpsertSourceByID(ctx, db.UpsertSourceByIDParams{ID: postgresUUID(v.ID), Name: v.Name, Kind: db.SourceKind(v.Kind), HomepageUrl: v.HomepageURL, LicenseName: v.LicenseName, LicenseUrl: v.LicenseURL, Attribution: v.Attribution, IsActive: v.IsActive})
		value = source(row)
	case "categories":
		v := input.(domain.CategoryUpsert)
		if byID {
			var id pgtype.UUID
			id, err = r.queries.UpsertCategoryByID(ctx, db.UpsertCategoryByIDParams{ID: postgresUUID(v.ID), ParentID: nullablePostgresUUID(v.ParentID), Slug: v.Slug, NameI18n: v.NameI18n})
			if err == nil {
				value, err = r.GetResource(ctx, resource, domain.ResourceKey{ID: uuidString(id)})
			}
		} else {
			var row db.Category
			row, err = r.queries.UpsertCategoryBySlug(ctx, db.UpsertCategoryBySlugParams{ParentID: nullablePostgresUUID(v.ParentID), Slug: v.Slug, NameI18n: v.NameI18n})
			value = category(row)
		}
	case "units":
		v := input.(domain.Unit)
		number, e := numericFromString(v.ToBaseFactor)
		if e != nil {
			return nil, e
		}
		var row db.Unit
		row, err = r.queries.UpsertUnitByCode(ctx, db.UpsertUnitByCodeParams{Code: v.Code, Dimension: v.Dimension, ToBaseFactor: number})
		value = unit(row)
	case "merchants":
		v := input.(domain.MerchantUpsert)
		var row db.Merchant
		row, err = r.queries.UpsertMerchantByID(ctx, db.UpsertMerchantByIDParams{ID: postgresUUID(v.ID), Name: v.Name, WebsiteUrl: v.WebsiteURL})
		value = merchant(row)
	case "locations":
		v := input.(domain.LocationUpsert)
		var row db.Location
		row, err = r.queries.UpsertLocationByID(ctx, db.UpsertLocationByIDParams{ID: postgresUUID(v.ID), MerchantID: nullablePostgresUUID(v.MerchantID), GeoAreaID: postgresUUID(v.GeoAreaID), Name: v.Name, Address: v.Address, Latitude: v.Latitude, Longitude: v.Longitude, ExternalRef: v.ExternalRef})
		if err == nil {
			value, err = r.GetResource(ctx, resource, domain.ResourceKey{ID: uuidString(row.ID)})
		}
	case "fuel-types":
		v := input.(domain.FuelTypeUpsert)
		if byID {
			var id pgtype.UUID
			id, err = r.queries.UpsertFuelTypeByID(ctx, db.UpsertFuelTypeByIDParams{ID: postgresUUID(v.ID), Code: v.Code, NameI18n: v.NameI18n, Energy: v.Energy})
			if err == nil {
				value, err = r.GetResource(ctx, resource, domain.ResourceKey{ID: uuidString(id)})
			}
		} else {
			var row db.FuelType
			row, err = r.queries.UpsertFuelTypeByCode(ctx, db.UpsertFuelTypeByCodeParams{Code: v.Code, NameI18n: v.NameI18n, Energy: v.Energy})
			value = fuelType(row)
		}
	case "fuel-prices":
		value, err = r.upsertFuelPrice(ctx, input.(domain.FuelPriceUpsert), byID)
	case "housing-observations":
		value, err = r.upsertHousing(ctx, input.(domain.HousingUpsert), byID)
	case "income-observations":
		value, err = r.upsertIncome(ctx, input.(domain.IncomeUpsert), byID)
	case "evidence-files":
		v := input.(domain.EvidenceUpsert)
		var row db.EvidenceFile
		if byID {
			row, err = r.queries.UpsertEvidenceByID(ctx, db.UpsertEvidenceByIDParams{ID: postgresUUID(v.ID), ObservationType: v.ObservationType, ObservationID: postgresUUID(v.ObservationID), ObjectKey: v.ObjectKey, MediaType: v.MediaType, Sha256: v.SHA256, UploadedBy: nullablePostgresUUID(v.UploadedBy)})
		} else {
			row, err = r.queries.UpsertEvidenceByObjectKey(ctx, db.UpsertEvidenceByObjectKeyParams{ObservationType: v.ObservationType, ObservationID: postgresUUID(v.ObservationID), ObjectKey: v.ObjectKey, MediaType: v.MediaType, Sha256: v.SHA256, UploadedBy: nullablePostgresUUID(v.UploadedBy)})
		}
		value = evidence(row)
	case "price-aggregates":
		v := input.(domain.PriceAggregate)
		median, _ := numericFromString(v.MedianAmount)
		min, _ := numericFromString(v.MinAmount)
		max, _ := numericFromString(v.MaxAmount)
		confidence, _ := numericFromString(v.ConfidenceScore)
		row, e := r.queries.UpsertMonthlyPriceAggregate(ctx, db.UpsertMonthlyPriceAggregateParams{GeoAreaID: postgresUUID(v.GeoAreaID), MetricType: v.MetricType, SubjectID: postgresUUID(v.SubjectID), Month: postgresDate(v.Month), MedianAmount: median, MinAmount: min, MaxAmount: max, SampleSize: v.SampleSize, ConfidenceScore: confidence, CalculatedAt: pgtype.Timestamptz{Time: v.CalculatedAt, Valid: true}})
		err = e
		value = aggregate(row)
	case "moderation-events":
		v := input.(domain.ModerationEventAppend)
		var previous *db.ModerationStatus
		if v.PreviousStatus != nil {
			status := db.ModerationStatus(*v.PreviousStatus)
			previous = &status
		}
		row, e := r.queries.AppendModerationEvent(ctx, db.AppendModerationEventParams{ObservationType: v.ObservationType, ObservationID: postgresUUID(v.ObservationID), ModeratorID: nullablePostgresUUID(v.ModeratorID), PreviousStatus: previous, NewStatus: db.ModerationStatus(v.NewStatus), ReasonCode: v.ReasonCode, Note: v.Note})
		err = e
		value = moderation(row)
	default:
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, mapWriteError(err)
	}
	return value, nil
}

func (r *PostgreSQL) DeleteResource(ctx context.Context, resource string, key domain.ResourceKey) error {
	var rows int64
	var err error
	switch resource {
	case "geo-areas":
		rows, err = r.queries.DeleteGeoArea(ctx, postgresUUID(key.ID))
	case "sources":
		rows, err = r.queries.DeleteSource(ctx, postgresUUID(key.ID))
	case "categories":
		rows, err = r.queries.DeleteCategory(ctx, postgresUUID(key.ID))
	case "units":
		rows, err = r.queries.DeleteUnit(ctx, key.Code)
	case "merchants":
		rows, err = r.queries.DeleteMerchant(ctx, postgresUUID(key.ID))
	case "locations":
		rows, err = r.queries.DeleteLocation(ctx, postgresUUID(key.ID))
	case "fuel-types":
		rows, err = r.queries.DeleteFuelType(ctx, postgresUUID(key.ID))
	case "fuel-prices":
		rows, err = r.queries.DeleteFuelPrice(ctx, postgresUUID(key.ID))
	case "housing-observations":
		rows, err = r.queries.DeleteHousingObservation(ctx, postgresUUID(key.ID))
	case "income-observations":
		rows, err = r.queries.DeleteIncomeObservation(ctx, postgresUUID(key.ID))
	case "evidence-files":
		rows, err = r.queries.DeleteEvidenceFile(ctx, postgresUUID(key.ID))
	case "price-aggregates":
		rows, err = r.queries.DeletePriceAggregate(ctx, db.DeletePriceAggregateParams{GeoAreaID: postgresUUID(key.GeoAreaID), MetricType: key.MetricType, SubjectID: postgresUUID(key.SubjectID), Month: postgresDate(key.Month)})
	default:
		return domain.ErrNotFound
	}
	if err != nil {
		return mapWriteError(err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PostgreSQL) upsertFuelPrice(ctx context.Context, v domain.FuelPriceUpsert, byID bool) (domain.FuelPrice, error) {
	amount, _ := numericFromString(v.AmountPerLitre)
	confidence, _ := nullableNumeric(v.ConfidenceScore)
	base := db.UpsertFuelPriceBySourceRecordParams{FuelTypeID: postgresUUID(v.FuelTypeID), LocationID: postgresUUID(v.LocationID), SourceID: postgresUUID(v.SourceID), ContributorID: nullablePostgresUUID(v.ContributorID), AmountPerLitre: amount, Currency: strings.ToUpper(v.Currency), ObservedAt: pgtype.Timestamptz{Time: v.ObservedAt, Valid: true}, Status: db.ModerationStatus(v.Status), ConfidenceScore: confidence}
	var row db.FuelPriceObservation
	var err error
	if byID {
		row, err = r.queries.UpsertFuelPriceByID(ctx, db.UpsertFuelPriceByIDParams{ID: postgresUUID(v.ID), FuelTypeID: base.FuelTypeID, LocationID: base.LocationID, SourceID: base.SourceID, ContributorID: base.ContributorID, AmountPerLitre: base.AmountPerLitre, Currency: base.Currency, ObservedAt: base.ObservedAt, Status: base.Status, ConfidenceScore: base.ConfidenceScore, SourceRecordID: v.SourceRecordID})
	} else {
		base.SourceRecordID = *v.SourceRecordID
		row, err = r.queries.UpsertFuelPriceBySourceRecord(ctx, base)
	}
	return fuelPrice(row), err
}
func (r *PostgreSQL) upsertHousing(ctx context.Context, v domain.HousingUpsert, byID bool) (domain.HousingObservation, error) {
	amount, _ := numericFromString(v.Amount)
	per, _ := nullableNumeric(v.AmountPerSqm)
	surface, _ := nullableNumeric(v.SurfaceSqm)
	base := db.UpsertHousingBySourceRecordParams{GeoAreaID: postgresUUID(v.GeoAreaID), SourceID: postgresUUID(v.SourceID), TransactionType: db.HousingTransaction(v.TransactionType), PropertyType: v.PropertyType, Rooms: v.Rooms, Furnished: v.Furnished, Amount: amount, AmountPerSqm: per, Currency: strings.ToUpper(v.Currency), SurfaceSqm: surface, ObservedAt: pgtype.Timestamptz{Time: v.ObservedAt, Valid: true}, Status: db.ModerationStatus(v.Status)}
	var row db.HousingObservation
	var err error
	if byID {
		row, err = r.queries.UpsertHousingByID(ctx, db.UpsertHousingByIDParams{ID: postgresUUID(v.ID), GeoAreaID: base.GeoAreaID, SourceID: base.SourceID, TransactionType: base.TransactionType, PropertyType: base.PropertyType, Rooms: base.Rooms, Furnished: base.Furnished, Amount: base.Amount, AmountPerSqm: base.AmountPerSqm, Currency: base.Currency, SurfaceSqm: base.SurfaceSqm, ObservedAt: base.ObservedAt, Status: base.Status, SourceRecordID: v.SourceRecordID})
	} else {
		base.SourceRecordID = *v.SourceRecordID
		row, err = r.queries.UpsertHousingBySourceRecord(ctx, base)
	}
	return housing(row), err
}
func (r *PostgreSQL) upsertIncome(ctx context.Context, v domain.IncomeUpsert, byID bool) (domain.IncomeObservation, error) {
	amount, _ := numericFromString(v.Amount)
	base := db.UpsertIncomeBySourceRecordParams{GeoAreaID: postgresUUID(v.GeoAreaID), SourceID: postgresUUID(v.SourceID), OccupationCode: v.OccupationCode, IndustryCode: v.IndustryCode, Amount: amount, Currency: strings.ToUpper(v.Currency), Period: db.IncomePeriod(v.Period), Basis: db.IncomeBasis(v.Basis), Statistic: v.Statistic, SampleSize: v.SampleSize, PeriodStart: postgresDate(v.PeriodStart), PeriodEnd: postgresDate(v.PeriodEnd), Status: db.ModerationStatus(v.Status)}
	var row db.IncomeObservation
	var err error
	if byID {
		row, err = r.queries.UpsertIncomeByID(ctx, db.UpsertIncomeByIDParams{ID: postgresUUID(v.ID), GeoAreaID: base.GeoAreaID, SourceID: base.SourceID, OccupationCode: base.OccupationCode, IndustryCode: base.IndustryCode, Amount: base.Amount, Currency: base.Currency, Period: base.Period, Basis: base.Basis, Statistic: base.Statistic, SampleSize: base.SampleSize, PeriodStart: base.PeriodStart, PeriodEnd: base.PeriodEnd, Status: base.Status, SourceRecordID: v.SourceRecordID})
	} else {
		base.SourceRecordID = *v.SourceRecordID
		row, err = r.queries.UpsertIncomeBySourceRecord(ctx, base)
	}
	return income(row), err
}

func postgresDate(value string) pgtype.Date {
	parsed, _ := time.Parse("2006-01-02", value)
	return pgtype.Date{Time: parsed, Valid: true}
}
func optionalUUID(value pgtype.UUID) *string {
	if !value.Valid {
		return nil
	}
	text := uuidString(value)
	return &text
}
func optionalTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
func optionalCoordinate(value any) *float64 {
	var number float64
	switch typed := value.(type) {
	case float64:
		number = typed
	case []byte:
		number, _ = strconv.ParseFloat(string(typed), 64)
	case string:
		number, _ = strconv.ParseFloat(typed, 64)
	default:
		return nil
	}
	if math.IsNaN(number) {
		return nil
	}
	return &number
}
func geoArea(id, parent pgtype.UUID, kind string, code *string, name, country string, latitude, longitude any) domain.GeoArea {
	return domain.GeoArea{ID: uuidString(id), ParentID: optionalUUID(parent), Type: kind, Code: code, Name: name, CountryCode: strings.TrimSpace(country), Latitude: optionalCoordinate(latitude), Longitude: optionalCoordinate(longitude)}
}
func source(v db.Source) domain.Source {
	return domain.Source{ID: uuidString(v.ID), Name: v.Name, Kind: string(v.Kind), HomepageURL: v.HomepageUrl, LicenseName: v.LicenseName, LicenseURL: v.LicenseUrl, Attribution: v.Attribution, IsActive: v.IsActive, CreatedAt: v.CreatedAt.Time}
}
func category(v db.Category) domain.Category {
	return domain.Category{ID: uuidString(v.ID), ParentID: optionalUUID(v.ParentID), Slug: v.Slug, NameI18n: validJSON(v.NameI18n, "{}")}
}
func unit(v db.Unit) domain.Unit {
	return domain.Unit{Code: v.Code, Dimension: v.Dimension, ToBaseFactor: numericString(v.ToBaseFactor)}
}
func merchant(v db.Merchant) domain.Merchant {
	return domain.Merchant{ID: uuidString(v.ID), Name: v.Name, WebsiteURL: v.WebsiteUrl, CreatedAt: v.CreatedAt.Time}
}
func location(id, merchantID, geoID pgtype.UUID, name string, address, external *string, created pgtype.Timestamptz, latitude, longitude any) domain.Location {
	return domain.Location{ID: uuidString(id), MerchantID: optionalUUID(merchantID), GeoAreaID: uuidString(geoID), Name: name, Address: address, ExternalRef: external, CreatedAt: created.Time, Latitude: optionalCoordinate(latitude), Longitude: optionalCoordinate(longitude)}
}
func fuelType(v db.FuelType) domain.FuelType {
	return domain.FuelType{ID: uuidString(v.ID), Code: v.Code, NameI18n: validJSON(v.NameI18n, "{}"), Energy: v.Energy}
}
func fuelPrice(v db.FuelPriceObservation) domain.FuelPrice {
	return domain.FuelPrice{ID: uuidString(v.ID), FuelTypeID: uuidString(v.FuelTypeID), LocationID: uuidString(v.LocationID), SourceID: uuidString(v.SourceID), ContributorID: optionalUUID(v.ContributorID), AmountPerLitre: numericString(v.AmountPerLitre), Currency: strings.TrimSpace(v.Currency), ObservedAt: v.ObservedAt.Time, Status: string(v.Status), ConfidenceScore: nullableNumericString(v.ConfidenceScore), SourceRecordID: v.SourceRecordID, CreatedAt: v.CreatedAt.Time}
}
func housing(v db.HousingObservation) domain.HousingObservation {
	return domain.HousingObservation{ID: uuidString(v.ID), GeoAreaID: uuidString(v.GeoAreaID), SourceID: uuidString(v.SourceID), TransactionType: string(v.TransactionType), PropertyType: v.PropertyType, Rooms: v.Rooms, Furnished: v.Furnished, Amount: numericString(v.Amount), AmountPerSqm: nullableNumericString(v.AmountPerSqm), Currency: strings.TrimSpace(v.Currency), SurfaceSqm: nullableNumericString(v.SurfaceSqm), ObservedAt: v.ObservedAt.Time, Status: string(v.Status), SourceRecordID: v.SourceRecordID, CreatedAt: v.CreatedAt.Time}
}
func income(v db.IncomeObservation) domain.IncomeObservation {
	return domain.IncomeObservation{ID: uuidString(v.ID), GeoAreaID: uuidString(v.GeoAreaID), SourceID: uuidString(v.SourceID), OccupationCode: v.OccupationCode, IndustryCode: v.IndustryCode, Amount: numericString(v.Amount), Currency: strings.TrimSpace(v.Currency), Period: string(v.Period), Basis: string(v.Basis), Statistic: v.Statistic, SampleSize: v.SampleSize, PeriodStart: v.PeriodStart.Time.Format("2006-01-02"), PeriodEnd: v.PeriodEnd.Time.Format("2006-01-02"), Status: string(v.Status), SourceRecordID: v.SourceRecordID, CreatedAt: v.CreatedAt.Time}
}
func evidence(v db.EvidenceFile) domain.EvidenceFile {
	return domain.EvidenceFile{ID: uuidString(v.ID), ObservationType: v.ObservationType, ObservationID: uuidString(v.ObservationID), ObjectKey: v.ObjectKey, MediaType: v.MediaType, SHA256: strings.TrimSpace(v.Sha256), UploadedBy: optionalUUID(v.UploadedBy), CreatedAt: v.CreatedAt.Time}
}
func aggregate(v db.PriceAggregatesMonthly) domain.PriceAggregate {
	return domain.PriceAggregate{GeoAreaID: uuidString(v.GeoAreaID), MetricType: v.MetricType, SubjectID: uuidString(v.SubjectID), Month: v.Month.Time.Format("2006-01-02"), MedianAmount: numericString(v.MedianAmount), MinAmount: numericString(v.MinAmount), MaxAmount: numericString(v.MaxAmount), SampleSize: v.SampleSize, ConfidenceScore: numericString(v.ConfidenceScore), CalculatedAt: v.CalculatedAt.Time}
}
func moderation(v db.ModerationEvent) domain.ModerationEvent {
	var previous *string
	if v.PreviousStatus != nil {
		text := string(*v.PreviousStatus)
		previous = &text
	}
	return domain.ModerationEvent{ID: uuidString(v.ID), ObservationType: v.ObservationType, ObservationID: uuidString(v.ObservationID), ModeratorID: optionalUUID(v.ModeratorID), PreviousStatus: previous, NewStatus: string(v.NewStatus), ReasonCode: v.ReasonCode, Note: v.Note, CreatedAt: v.CreatedAt.Time}
}
func adminUser(id pgtype.UUID, email string, display *string, role db.UserRole, locale string, reputation int32, verified, created, updated, deleted pgtype.Timestamptz) domain.AdminUser {
	return domain.AdminUser{ID: uuidString(id), Email: email, DisplayName: display, Role: string(role), Locale: locale, ReputationScore: reputation, EmailVerifiedAt: optionalTime(verified), CreatedAt: created.Time, UpdatedAt: updated.Time, DeletedAt: optionalTime(deleted)}
}
