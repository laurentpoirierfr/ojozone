package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/store/db"
)

// FindCommunitySourceID retourne la source citoyenne référencée par les contributions.
func (r *PostgreSQL) FindCommunitySourceID(ctx context.Context) (string, error) {
	id, err := r.queries.GetCommunitySourceID(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return uuidString(id), nil
}

// CreateProductContribution insère un prix proposé avec le statut pending.
func (r *PostgreSQL) CreateProductContribution(ctx context.Context, submit domain.ProductContributionSubmit) (string, string, error) {
	amount, err := numericFromString(submit.Amount)
	if err != nil {
		return "", "", err
	}
	quantity, err := numericFromString(submit.Quantity)
	if err != nil {
		return "", "", err
	}
	normalizedAmount, err := numericFromString(submit.NormalizedAmount)
	if err != nil {
		return "", "", err
	}
	observations, err := r.queries.InsertProductContribution(ctx, db.InsertProductContributionParams{
		ProductID:        postgresUUID(submit.ProductID),
		LocationID:       postgresUUID(submit.LocationID),
		SourceID:         postgresUUID(submit.SourceID),
		ContributorID:    postgresUUID(submit.ContributorID),
		Amount:           amount,
		Currency:         submit.Currency,
		Quantity:         quantity,
		UnitCode:         submit.UnitCode,
		NormalizedAmount: normalizedAmount,
		IsPromotion:      submit.IsPromotion,
		ObservedAt:       pgtype.Timestamptz{Time: submit.ObservedAt, Valid: true},
	})
	if err != nil {
		return "", "", mapWriteError(err)
	}
	return uuidString(observations.ID), string(observations.Status), nil
}

// CreateFuelContribution insère un prix carburant proposé avec le statut pending.
func (r *PostgreSQL) CreateFuelContribution(ctx context.Context, submit domain.FuelContributionSubmit) (string, string, error) {
	amountPerLitre, err := numericFromString(submit.AmountPerLitre)
	if err != nil {
		return "", "", err
	}
	observations, err := r.queries.InsertFuelContribution(ctx, db.InsertFuelContributionParams{
		FuelTypeID:     postgresUUID(submit.FuelTypeID),
		LocationID:     postgresUUID(submit.LocationID),
		SourceID:       postgresUUID(submit.SourceID),
		ContributorID:  postgresUUID(submit.ContributorID),
		AmountPerLitre: amountPerLitre,
		Currency:       submit.Currency,
		ObservedAt:     pgtype.Timestamptz{Time: submit.ObservedAt, Valid: true},
	})
	if err != nil {
		return "", "", mapWriteError(err)
	}
	return uuidString(observations.ID), string(observations.Status), nil
}

// GetProductContribution retourne une contribution produit par UUID.
func (r *PostgreSQL) GetProductContribution(ctx context.Context, id string) (domain.ContributionDetail, error) {
	row, err := r.queries.GetProductContribution(ctx, postgresUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ContributionDetail{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ContributionDetail{}, err
	}
	return productContributionDetail(row), nil
}

// GetFuelContribution retourne une contribution carburant par UUID.
func (r *PostgreSQL) GetFuelContribution(ctx context.Context, id string) (domain.ContributionDetail, error) {
	row, err := r.queries.GetFuelContribution(ctx, postgresUUID(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ContributionDetail{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ContributionDetail{}, err
	}
	return fuelContributionDetail(row), nil
}

// UpdateProductContribution corrige une contribution produit encore pending.
func (r *PostgreSQL) UpdateProductContribution(ctx context.Context, id, contributorID string, submit domain.ProductContributionSubmit) error {
	amount, err := numericFromString(submit.Amount)
	if err != nil {
		return err
	}
	quantity, err := numericFromString(submit.Quantity)
	if err != nil {
		return err
	}
	normalizedAmount, err := numericFromString(submit.NormalizedAmount)
	if err != nil {
		return err
	}
	_, err = r.queries.UpdateProductContribution(ctx, db.UpdateProductContributionParams{
		ID:               postgresUUID(id),
		ContributorID:    postgresUUID(contributorID),
		Amount:           amount,
		Currency:         submit.Currency,
		Quantity:         quantity,
		UnitCode:         submit.UnitCode,
		NormalizedAmount: normalizedAmount,
		IsPromotion:      submit.IsPromotion,
		ObservedAt:       pgtype.Timestamptz{Time: submit.ObservedAt, Valid: true},
	})
	if err != nil {
		return err
	}
	return nil
}

// UpdateFuelContribution corrige une contribution carburant encore pending.
func (r *PostgreSQL) UpdateFuelContribution(ctx context.Context, id, contributorID string, submit domain.FuelContributionSubmit) error {
	amountPerLitre, err := numericFromString(submit.AmountPerLitre)
	if err != nil {
		return err
	}
	_, err = r.queries.UpdateFuelContribution(ctx, db.UpdateFuelContributionParams{
		ID:             postgresUUID(id),
		ContributorID:  postgresUUID(contributorID),
		AmountPerLitre: amountPerLitre,
		Currency:       submit.Currency,
		ObservedAt:     pgtype.Timestamptz{Time: submit.ObservedAt, Valid: true},
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteProductContribution retire une contribution produit encore pending.
func (r *PostgreSQL) DeleteProductContribution(ctx context.Context, id, contributorID string) error {
	rowsAffected, err := r.queries.DeleteProductContribution(ctx, db.DeleteProductContributionParams{
		ID: postgresUUID(id), ContributorID: postgresUUID(contributorID),
	})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DeleteFuelContribution retire une contribution carburant encore pending.
func (r *PostgreSQL) DeleteFuelContribution(ctx context.Context, id, contributorID string) error {
	rowsAffected, err := r.queries.DeleteFuelContribution(ctx, db.DeleteFuelContributionParams{
		ID: postgresUUID(id), ContributorID: postgresUUID(contributorID),
	})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func productContributionDetail(row db.GetProductContributionRow) domain.ContributionDetail {
	quantity := numericString(row.Quantity)
	unit := strings.TrimSpace(row.UnitCode)
	isPromotion := row.IsPromotion
	return domain.ContributionDetail{
		Contribution: domain.Contribution{
			ID: uuidString(row.ID), Type: "product_price", Status: string(row.Status),
			ObservedAt: row.ObservedAt.Time, CreatedAt: row.CreatedAt.Time,
			Currency: strings.TrimSpace(row.Currency), Amount: numericString(row.Amount),
			Quantity: &quantity, UnitCode: &unit, IsPromotion: &isPromotion,
			ConfidenceScore: nullableNumericString(row.ConfidenceScore), Subject: row.ProductName,
			Location: domain.EntityRef{ID: uuidString(row.LocationID), Name: row.LocationName},
			GeoArea:  domain.EntityRef{ID: uuidString(row.GeoAreaID), Name: row.GeoAreaName},
			Source:   domain.SourceRef{ID: uuidString(row.SourceID), Name: row.SourceName, Kind: string(row.SourceKind)},
		},
		ContributorID: uuidString(row.ContributorID),
	}
}

func fuelContributionDetail(row db.GetFuelContributionRow) domain.ContributionDetail {
	return domain.ContributionDetail{
		Contribution: domain.Contribution{
			ID: uuidString(row.ID), Type: "fuel_price", Status: string(row.Status),
			ObservedAt: row.ObservedAt.Time, CreatedAt: row.CreatedAt.Time,
			Currency: strings.TrimSpace(row.Currency), Amount: numericString(row.AmountPerLitre),
			ConfidenceScore: nullableNumericString(row.ConfidenceScore), Subject: row.FuelTypeCode,
			Location: domain.EntityRef{ID: uuidString(row.LocationID), Name: row.LocationName},
			GeoArea:  domain.EntityRef{ID: uuidString(row.GeoAreaID), Name: row.GeoAreaName},
			Source:   domain.SourceRef{ID: uuidString(row.SourceID), Name: row.SourceName, Kind: string(row.SourceKind)},
		},
		ContributorID: uuidString(row.ContributorID),
	}
}
