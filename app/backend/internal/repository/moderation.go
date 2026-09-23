package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/laurentpoirierfr/ojozone/internal/domain"
	"github.com/laurentpoirierfr/ojozone/internal/store/db"
)

// ListModerationQueue retourne les contributions de la file, toutes natures
// confondues, triées de la plus récente à la plus ancienne.
func (r *PostgreSQL) ListModerationQueue(ctx context.Context, status *string, pagination domain.Pagination) ([]domain.ModerationQueueItem, error) {
	rows, err := r.queries.ListModerationQueue(ctx, db.ListModerationQueueParams{
		Status: status, PageSize: pagination.Limit, PageOffset: pagination.Offset,
	})
	if err != nil {
		return nil, err
	}
	items := make([]domain.ModerationQueueItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, moderationQueueItem(row))
	}
	return items, nil
}

// ReviewContribution applique une décision (approbation ou rejet) de façon
// atomique : mise à jour du statut puis événement de modération append-only.
func (r *PostgreSQL) ReviewContribution(ctx context.Context, review domain.ContributionReview) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	transactions := r.queries.WithTx(tx)
	var previous db.ModerationStatus
	switch review.Type {
	case "product_price":
		row, err := transactions.GetProductContribution(ctx, postgresUUID(review.ID))
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		previous = row.Status
		rows, err := transactions.SetProductContributionStatus(ctx, db.SetProductContributionStatusParams{
			ID:     postgresUUID(review.ID),
			Status: db.ModerationStatus(review.NewStatus),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return domain.ErrContributionNotModifiable
		}
	case "fuel_price":
		row, err := transactions.GetFuelContribution(ctx, postgresUUID(review.ID))
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		previous = row.Status
		rows, err := transactions.SetFuelContributionStatus(ctx, db.SetFuelContributionStatusParams{
			ID:     postgresUUID(review.ID),
			Status: db.ModerationStatus(review.NewStatus),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return domain.ErrContributionNotModifiable
		}
	default:
		return domain.ErrNotFound
	}

	if _, err := transactions.AppendModerationEvent(ctx, db.AppendModerationEventParams{
		ObservationType: review.Type,
		ObservationID:   postgresUUID(review.ID),
		ModeratorID:     postgresUUID(review.ModeratorID),
		PreviousStatus:  &previous,
		NewStatus:       db.ModerationStatus(review.NewStatus),
		ReasonCode:      review.ReasonCode,
		Note:            review.Note,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func moderationQueueItem(row db.ListModerationQueueRow) domain.ModerationQueueItem {
	return domain.ModerationQueueItem{
		ID:               uuidString(row.ID),
		Type:             row.ObservationType,
		Status:           string(row.Status),
		Subject:          fmt.Sprint(row.Subject),
		ContributorID:    optionalUUID(row.ContributorID),
		ContributorEmail: row.ContributorEmail,
		Amount:           row.Amount,
		Currency:         strings.TrimSpace(row.Currency),
		Quantity:         nullableText(row.Quantity),
		UnitCode:         nullableText(row.UnitCode),
		IsPromotion:      row.IsPromotion,
		Location:         domain.EntityRef{ID: uuidString(row.LocationID), Name: row.LocationName},
		GeoArea:          domain.EntityRef{ID: uuidString(row.GeoAreaID), Name: row.GeoAreaName},
		Source:           domain.SourceRef{ID: uuidString(row.SourceID), Name: row.SourceName, Kind: string(row.SourceKind)},
		ObservedAt:       row.ObservedAt.Time,
		ConfidenceScore:  nullableNumericString(row.ConfidenceScore),
		CreatedAt:        row.CreatedAt.Time,
	}
}

// nullableText normalise une chaîne nullable : pointeur vers la valeur trimée,
// ou nil si absente ou vide.
func nullableText(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
