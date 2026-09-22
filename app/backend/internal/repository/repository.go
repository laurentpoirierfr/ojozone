package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/store/db"
)

type Repository interface {
	Ping(context.Context) error
	CreateUser(context.Context, domain.RegisterInput, string) (domain.User, error)
	GetUserByEmail(context.Context, string) (domain.UserWithPassword, error)
	GetUserByID(context.Context, string) (domain.User, error)
	UpdateUserProfile(context.Context, string, domain.UpdateProfileInput) (domain.User, error)
	AnonymizeUser(context.Context, string) (domain.User, error)
	CreateSession(context.Context, domain.Session) (domain.Session, error)
	GetSessionByRefreshHash(context.Context, string) (domain.Session, error)
	UpdateSessionRefresh(context.Context, string, string, time.Time) (domain.Session, error)
	RevokeSession(context.Context, string) error
	RevokeAllSessionsForUser(context.Context, string) error
	ListMyContributions(context.Context, string, domain.Pagination) ([]domain.Contribution, error)
	FindCommunitySourceID(context.Context) (string, error)
	CreateProductContribution(context.Context, domain.ProductContributionSubmit) (string, string, error)
	CreateFuelContribution(context.Context, domain.FuelContributionSubmit) (string, string, error)
	GetProductContribution(context.Context, string) (domain.ContributionDetail, error)
	GetFuelContribution(context.Context, string) (domain.ContributionDetail, error)
	UpdateProductContribution(context.Context, string, string, domain.ProductContributionSubmit) error
	UpdateFuelContribution(context.Context, string, string, domain.FuelContributionSubmit) error
	DeleteProductContribution(context.Context, string, string) error
	DeleteFuelContribution(context.Context, string, string) error
	ListProducts(context.Context, domain.ProductFilter) ([]domain.Product, error)
	GetProduct(context.Context, string) (domain.Product, error)
	GetProductByBarcode(context.Context, string) (domain.Product, error)
	UpsertProductByID(context.Context, domain.ProductUpsert) (domain.Product, error)
	UpsertProductByBarcode(context.Context, domain.ProductUpsert) (domain.Product, error)
	DeleteProduct(context.Context, string) error
	ListProductPrices(context.Context, string, domain.PriceFilter) ([]domain.ProductPrice, error)
	ListAllProductPrices(context.Context, domain.PriceFilter) ([]domain.ProductPrice, error)
	GetProductPrice(context.Context, string) (domain.ProductPrice, error)
	UpsertProductPriceBySourceRecord(context.Context, domain.ProductPriceUpsert) (domain.ProductPrice, error)
	UpsertProductPriceByID(context.Context, domain.ProductPriceUpsert) (domain.ProductPrice, error)
	DeleteProductPrice(context.Context, string) error
	ListResource(context.Context, string, domain.Pagination) (any, error)
	GetResource(context.Context, string, domain.ResourceKey) (any, error)
	UpsertResource(context.Context, string, any, bool) (any, error)
	DeleteResource(context.Context, string, domain.ResourceKey) error
}

type PostgreSQL struct {
	queries *db.Queries
}

func NewPostgreSQL(queries *db.Queries) *PostgreSQL {
	return &PostgreSQL{queries: queries}
}

func (r *PostgreSQL) Ping(ctx context.Context) error {
	_, err := r.queries.Ping(ctx)
	return err
}

func (r *PostgreSQL) ListProducts(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	var search *string
	if filter.Search != "" {
		search = &filter.Search
	}
	rows, err := r.queries.ListProducts(ctx, db.ListProductsParams{
		Search: search, PageSize: filter.Limit, PageOffset: filter.Offset,
	})
	if err != nil {
		return nil, err
	}
	products := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, productFromListRow(row))
	}
	return products, nil
}

func (r *PostgreSQL) GetProduct(ctx context.Context, id string) (domain.Product, error) {
	row, err := r.queries.GetProduct(ctx, postgresUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Product{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Product{}, err
	}
	return productFromGetRow(row), nil
}

func (r *PostgreSQL) GetProductByBarcode(ctx context.Context, barcode string) (domain.Product, error) {
	row, err := r.queries.GetProductByBarcode(ctx, barcode)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Product{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Product{}, err
	}
	return productFromBarcodeRow(row), nil
}

func (r *PostgreSQL) UpsertProductByID(ctx context.Context, input domain.ProductUpsert) (domain.Product, error) {
	row, err := r.queries.UpsertProductByID(ctx, db.UpsertProductByIDParams{
		ID: postgresUUID(input.ID), CategoryID: postgresUUID(input.CategoryID), Name: input.Name,
		Brand: input.Brand, Barcode: input.Barcode, ReferenceUnitCode: input.ReferenceUnit,
		IsGeneric: input.IsGeneric, Attributes: input.Attributes,
	})
	if err != nil {
		return domain.Product{}, mapWriteError(err)
	}
	return r.GetProduct(ctx, uuidString(row.ID))
}

func (r *PostgreSQL) UpsertProductByBarcode(ctx context.Context, input domain.ProductUpsert) (domain.Product, error) {
	row, err := r.queries.UpsertProductByBarcode(ctx, db.UpsertProductByBarcodeParams{
		CategoryID: postgresUUID(input.CategoryID), Name: input.Name, Brand: input.Brand,
		Barcode: *input.Barcode, ReferenceUnitCode: input.ReferenceUnit,
		IsGeneric: input.IsGeneric, Attributes: input.Attributes,
	})
	if err != nil {
		return domain.Product{}, mapWriteError(err)
	}
	return r.GetProduct(ctx, uuidString(row.ID))
}

func (r *PostgreSQL) DeleteProduct(ctx context.Context, id string) error {
	rowsAffected, err := r.queries.DeleteProduct(ctx, postgresUUID(id))
	if err != nil {
		return mapWriteError(err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PostgreSQL) ListProductPrices(ctx context.Context, productID string, filter domain.PriceFilter) ([]domain.ProductPrice, error) {
	var geoAreaID pgtype.UUID
	if filter.GeoAreaID != "" {
		geoAreaID = postgresUUID(filter.GeoAreaID)
	}
	rows, err := r.queries.ListApprovedProductPrices(ctx, db.ListApprovedProductPricesParams{
		ProductID: postgresUUID(productID), GeoAreaID: geoAreaID,
		PageSize: filter.Limit, PageOffset: filter.Offset,
	})
	if err != nil {
		return nil, err
	}
	prices := make([]domain.ProductPrice, 0, len(rows))
	for _, row := range rows {
		prices = append(prices, priceFromRow(row))
	}
	return prices, nil
}

func (r *PostgreSQL) ListAllProductPrices(ctx context.Context, filter domain.PriceFilter) ([]domain.ProductPrice, error) {
	var productID pgtype.UUID
	if filter.ProductID != "" {
		productID = postgresUUID(filter.ProductID)
	}
	var geoAreaID pgtype.UUID
	if filter.GeoAreaID != "" {
		geoAreaID = postgresUUID(filter.GeoAreaID)
	}
	var status *string
	if filter.Status != "" {
		status = &filter.Status
	}
	rows, err := r.queries.ListProductPriceObservations(ctx, db.ListProductPriceObservationsParams{
		ProductID: productID, GeoAreaID: geoAreaID, Status: status,
		PageSize: filter.Limit, PageOffset: filter.Offset,
	})
	if err != nil {
		return nil, err
	}
	prices := make([]domain.ProductPrice, 0, len(rows))
	for _, row := range rows {
		prices = append(prices, priceFromObservationRow(row))
	}
	return prices, nil
}

func (r *PostgreSQL) GetProductPrice(ctx context.Context, id string) (domain.ProductPrice, error) {
	row, err := r.queries.GetProductPrice(ctx, postgresUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProductPrice{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ProductPrice{}, err
	}
	return priceFromGetRow(row), nil
}

func (r *PostgreSQL) UpsertProductPriceBySourceRecord(ctx context.Context, input domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	params, err := productPriceSourceParams(input)
	if err != nil {
		return domain.ProductPrice{}, err
	}
	row, err := r.queries.UpsertProductPriceBySourceRecord(ctx, params)
	if err != nil {
		return domain.ProductPrice{}, mapWriteError(err)
	}
	return r.GetProductPrice(ctx, uuidString(row.ID))
}

func (r *PostgreSQL) UpsertProductPriceByID(ctx context.Context, input domain.ProductPriceUpsert) (domain.ProductPrice, error) {
	amount, err := numericFromString(input.Amount)
	if err != nil {
		return domain.ProductPrice{}, err
	}
	quantity, err := numericFromString(input.Quantity)
	if err != nil {
		return domain.ProductPrice{}, err
	}
	normalizedAmount, err := numericFromString(input.NormalizedAmount)
	if err != nil {
		return domain.ProductPrice{}, err
	}
	confidenceScore, err := nullableNumeric(input.ConfidenceScore)
	if err != nil {
		return domain.ProductPrice{}, err
	}
	row, err := r.queries.UpsertProductPriceByID(ctx, db.UpsertProductPriceByIDParams{
		ID: postgresUUID(input.ID), ProductID: postgresUUID(input.ProductID),
		LocationID: postgresUUID(input.LocationID), SourceID: postgresUUID(input.SourceID),
		ContributorID: nullablePostgresUUID(input.ContributorID), Amount: amount,
		Currency: input.Currency, Quantity: quantity, UnitCode: input.UnitCode,
		NormalizedAmount: normalizedAmount, IsPromotion: input.IsPromotion,
		ObservedAt: pgtype.Timestamptz{Time: input.ObservedAt, Valid: true},
		Status:     db.ModerationStatus(input.Status), ConfidenceScore: confidenceScore,
		SourceRecordID: input.SourceRecordID,
	})
	if err != nil {
		return domain.ProductPrice{}, mapWriteError(err)
	}
	return r.GetProductPrice(ctx, uuidString(row.ID))
}

func (r *PostgreSQL) DeleteProductPrice(ctx context.Context, id string) error {
	rowsAffected, err := r.queries.DeleteProductPrice(ctx, postgresUUID(id))
	if err != nil {
		return mapWriteError(err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func postgresUUID(value string) pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.MustParse(value), Valid: true}
}

func nullablePostgresUUID(value *string) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}
	return postgresUUID(*value)
}

func numericFromString(value string) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric
	if err := numeric.Scan(value); err != nil {
		return pgtype.Numeric{}, err
	}
	return numeric, nil
}

func nullableNumeric(value *string) (pgtype.Numeric, error) {
	if value == nil {
		return pgtype.Numeric{}, nil
	}
	return numericFromString(*value)
}

func productPriceSourceParams(input domain.ProductPriceUpsert) (db.UpsertProductPriceBySourceRecordParams, error) {
	amount, err := numericFromString(input.Amount)
	if err != nil {
		return db.UpsertProductPriceBySourceRecordParams{}, err
	}
	quantity, err := numericFromString(input.Quantity)
	if err != nil {
		return db.UpsertProductPriceBySourceRecordParams{}, err
	}
	normalizedAmount, err := numericFromString(input.NormalizedAmount)
	if err != nil {
		return db.UpsertProductPriceBySourceRecordParams{}, err
	}
	confidenceScore, err := nullableNumeric(input.ConfidenceScore)
	if err != nil {
		return db.UpsertProductPriceBySourceRecordParams{}, err
	}
	return db.UpsertProductPriceBySourceRecordParams{
		ProductID: postgresUUID(input.ProductID), LocationID: postgresUUID(input.LocationID),
		SourceID: postgresUUID(input.SourceID), ContributorID: nullablePostgresUUID(input.ContributorID),
		Amount: amount, Currency: input.Currency, Quantity: quantity, UnitCode: input.UnitCode,
		NormalizedAmount: normalizedAmount, IsPromotion: input.IsPromotion,
		ObservedAt: pgtype.Timestamptz{Time: input.ObservedAt, Valid: true},
		Status:     db.ModerationStatus(input.Status), ConfidenceScore: confidenceScore,
		SourceRecordID: *input.SourceRecordID,
	}, nil
}

func mapWriteError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && (postgresError.Code == "23503" || postgresError.Code == "23505") {
		return domain.ErrConflict
	}
	return err
}

func uuidString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return uuid.UUID(value.Bytes).String()
}

func numericString(value pgtype.Numeric) string {
	result, err := value.Value()
	if err != nil || result == nil {
		return ""
	}
	return fmt.Sprint(result)
}

func nullableNumericString(value pgtype.Numeric) *string {
	if !value.Valid {
		return nil
	}
	result := numericString(value)
	return &result
}

func validJSON(value []byte, fallback string) json.RawMessage {
	if json.Valid(value) {
		return value
	}
	return json.RawMessage(fallback)
}

func productFromListRow(row db.ListProductsRow) domain.Product {
	return domain.Product{
		ID: uuidString(row.ID), Name: row.Name, Brand: row.Brand, Barcode: row.Barcode,
		ReferenceUnit: row.ReferenceUnitCode, IsGeneric: row.IsGeneric,
		Attributes: validJSON(row.Attributes, "{}"), CategorySlug: row.CategorySlug,
		CategoryNameI18n: validJSON(row.CategoryNameI18n, "{}"), CreatedAt: row.CreatedAt.Time,
	}
}

func productFromGetRow(row db.GetProductRow) domain.Product {
	return domain.Product{
		ID: uuidString(row.ID), Name: row.Name, Brand: row.Brand, Barcode: row.Barcode,
		ReferenceUnit: row.ReferenceUnitCode, IsGeneric: row.IsGeneric,
		Attributes: validJSON(row.Attributes, "{}"), CategorySlug: row.CategorySlug,
		CategoryNameI18n: validJSON(row.CategoryNameI18n, "{}"), CreatedAt: row.CreatedAt.Time,
	}
}

func productFromBarcodeRow(row db.GetProductByBarcodeRow) domain.Product {
	return domain.Product{
		ID: uuidString(row.ID), Name: row.Name, Brand: row.Brand, Barcode: row.Barcode,
		ReferenceUnit: row.ReferenceUnitCode, IsGeneric: row.IsGeneric,
		Attributes: validJSON(row.Attributes, "{}"), CategorySlug: row.CategorySlug,
		CategoryNameI18n: validJSON(row.CategoryNameI18n, "{}"), CreatedAt: row.CreatedAt.Time,
	}
}

func priceFromRow(row db.ListApprovedProductPricesRow) domain.ProductPrice {
	return domain.ProductPrice{
		ID: uuidString(row.ID), ProductID: uuidString(row.ProductID), Amount: numericString(row.Amount),
		Currency: strings.TrimSpace(row.Currency), Quantity: numericString(row.Quantity), UnitCode: row.UnitCode,
		NormalizedAmount: numericString(row.NormalizedAmount), IsPromotion: row.IsPromotion,
		ObservedAt: row.ObservedAt.Time, Status: string(row.Status),
		ConfidenceScore: nullableNumericString(row.ConfidenceScore), SourceRecordID: row.SourceRecordID,
		Location: domain.EntityRef{ID: uuidString(row.LocationID), Name: row.LocationName, Address: row.LocationAddress},
		GeoArea:  domain.EntityRef{ID: uuidString(row.GeoAreaID), Name: row.GeoAreaName},
		Source:   domain.SourceRef{ID: uuidString(row.SourceID), Name: row.SourceName, Kind: string(row.SourceKind)},
	}
}

func priceFromGetRow(row db.GetProductPriceRow) domain.ProductPrice {
	return domain.ProductPrice{
		ID: uuidString(row.ID), ProductID: uuidString(row.ProductID), Amount: numericString(row.Amount),
		Currency: strings.TrimSpace(row.Currency), Quantity: numericString(row.Quantity), UnitCode: row.UnitCode,
		NormalizedAmount: numericString(row.NormalizedAmount), IsPromotion: row.IsPromotion,
		ObservedAt: row.ObservedAt.Time, Status: string(row.Status),
		ConfidenceScore: nullableNumericString(row.ConfidenceScore), SourceRecordID: row.SourceRecordID,
		Location: domain.EntityRef{ID: uuidString(row.LocationID), Name: row.LocationName, Address: row.LocationAddress},
		GeoArea:  domain.EntityRef{ID: uuidString(row.GeoAreaID), Name: row.GeoAreaName},
		Source:   domain.SourceRef{ID: uuidString(row.SourceID), Name: row.SourceName, Kind: string(row.SourceKind)},
	}
}

func priceFromObservationRow(row db.ListProductPriceObservationsRow) domain.ProductPrice {
	return domain.ProductPrice{
		ID: uuidString(row.ID), ProductID: uuidString(row.ProductID), Amount: numericString(row.Amount),
		Currency: strings.TrimSpace(row.Currency), Quantity: numericString(row.Quantity), UnitCode: row.UnitCode,
		NormalizedAmount: numericString(row.NormalizedAmount), IsPromotion: row.IsPromotion,
		ObservedAt: row.ObservedAt.Time, Status: string(row.Status),
		ConfidenceScore: nullableNumericString(row.ConfidenceScore), SourceRecordID: row.SourceRecordID,
		Location: domain.EntityRef{ID: uuidString(row.LocationID), Name: row.LocationName, Address: row.LocationAddress},
		GeoArea:  domain.EntityRef{ID: uuidString(row.GeoAreaID), Name: row.GeoAreaName},
		Source:   domain.SourceRef{ID: uuidString(row.SourceID), Name: row.SourceName, Kind: string(row.SourceKind)},
	}
}
